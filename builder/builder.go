package builder

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"gosh_builder/sbom"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func Build(workdir string, cache bool) (target_sha string, err error) {
	target_sha_path := workdir + "/.targetsha"
	if cache {
		if _, err := os.Stat(target_sha_path); errors.Is(err, os.ErrNotExist) {
			out, err := os.ReadFile(target_sha_path)
			if err != nil {
				return "", err
			}

			target_sha = string(out)
			return target_sha, nil
		}
	}

	dockerfile_path := workdir + "/Dockerfile"
	df, err := os.ReadFile(dockerfile_path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			msg := fmt.Sprint("Dockerfile not found in", workdir)
			return "", errors.New(msg)
		} else {
			return "", err
		}
	}

	sbom_path := workdir + "/sbom.spdx.json"
	sbom_content, err := os.ReadFile(sbom_path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			msg := fmt.Sprint("SBOM not found in", workdir)
			return "", errors.New(msg)
		} else {
			return "", err
		}
	}
	sbom, err := sbom.ParseSBOM(sbom_content)
	if err != nil {
		return "", err
	}

	image_mapping := make(map[string]string)
	for _, component := range sbom.Components {
		if component.Type == "image" {
			// TODO: revert back the main logic
			// component_target_sha, err := goshAnytreeBuild(component.Purl)
			// if err != nil {
			// 	return "", err
			// }
			// image_mapping[component.Name] = component_target_sha
			image_mapping[component.Name] = component.Purl
		}
	}
	log.Println("image_lock_mapping", image_mapping)

	dockerfile_lock, err := dockerfileLock(df, image_mapping)
	if err != nil {
		return "", err
	}

	// make a new docker_file
	dockerfile_lock_path := workdir + "/lock.Dockerfile"
	if err := os.WriteFile(dockerfile_lock_path, []byte(dockerfile_lock), 0777); err != nil {
		return "", err
	}

	target_sha, err = dockerLockBuild(dockerfile_lock_path, workdir)
	if err != nil {
		return "", fmt.Errorf("failed to build docker image via lock.Dockerfile: %+v", err)
	}

	log.Println("Write docker image hash to", target_sha_path)
	if err = os.WriteFile(target_sha_path, []byte(target_sha), 0777); err != nil {
		return "", err
	}

	return target_sha, nil
}

func goshAnytreeBuild(purl string) (target_sha string, err error) {
	cmd := exec.Command("gosh", "anytree", "build", "--quiet", purl)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	target_sha = string(out)
	log.Println("target_sha", target_sha)
	return target_sha, err
}

func dockerLockBuild(dockerfile_lock_path, repo_dir string) (target_sha string, err error) {
	// TODO: exec.CommandContext() for more controll
	cmd := exec.Command(
		"docker", "buildx", "build",
		"-f", dockerfile_lock_path,
		repo_dir,
	)

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return target_sha, nil
}

func getCacheDir(repo_url string) (cache_dir string) {
	parts := strings.Split(repo_url, "/")
	hash := sha256.Sum256([]byte(repo_url))
	log.Println("hash", hash)

	cache_dir = fmt.Sprintf(
		"%s/.cache/gosh/builder/%s-%x",
		os.Getenv("HOME"),
		parts[len(parts)-1],
		hash,
	)
	log.Println("cache_dir", cache_dir)
	return cache_dir
}

func dockerfileLock(df []byte, image_mapping map[string]string) (dockerfile_lock string, err error) {
	dockerfile, err := parser.Parse(bytes.NewReader(df))
	if err != nil {
		return "", err
	}

	stages := []string{}

	// TODO: implement stage aliasing
	for _, node := range dockerfile.AST.Children {
		op_value := strings.ToUpper(node.Value)
		str := ""
		if op_value == "FROM" {
			str += op_value

			if node.Next == nil || node.Next.Value == "" {
				msg := "`Dockerfile` FROM is empty"
				return "", errors.New(msg)
			}

			if len(node.Flags) > 0 {
				str += dumpFlags(node.Flags)
			}

			image_name := node.Next.Value

			stage_index := getStageByAlias(image_name, stages)

			target_image, ok := image_mapping[image_name]

			if !ok && image_name != "scratch" && stage_index == -1 {
				msg := "image " + image_name + " not found"
				return "", errors.New(msg)
			}

			if ok && stage_index == -1 {
				str += " " + target_image
			} else {
				str += " " + image_name
			}

			if node.Next.Next == nil {
				stages = append(stages, "")
			} else {
				as_node := node.Next.Next
				if as_node.Value != "as" || as_node.Next == nil || as_node.Next.Value == "" {
					msg := "wrong `Dockerfile` FROM ... AS syntax"
					return "", errors.New(msg)
				}
				stages = append(stages, as_node.Next.Value)
				str += " as " + as_node.Next.Value
			}
		} else {
			str += naiveDump(node)
		}
		dockerfile_lock += str + "\n"
	}
	log.Println("stages", stages)
	return dockerfile_lock, nil
}

func dumpFlags(flags []string) (res string) {
	if len(flags) == 0 {
		return ""
	}
	if len(flags) == 1 {
		return " " + flags[0]
	}
	for _, flag := range flags {
		res += " \\\n" + "  " + flag
	}
	return res
}

func naiveDump(node *parser.Node) (res string) {
	if node == nil {
		return ""
	}
	res += node.Value

	// TODO: fix flags like `--from=<img>` and `--mount=type=cache`
	if len(node.Flags) > 0 {
		res += dumpFlags(node.Flags)
	}

	for n := node.Next; n != nil; n = n.Next {
		res += " " + n.Value
	}

	if len(node.Heredocs) > 0 {
		res += "\n" + node.Heredocs[0].Content + node.Heredocs[0].Name
	}

	return res
}

func getStageByAlias(stage_name string, stages []string) (stage_index int) {
	for stage_index = len(stages) - 1; stage_index >= 0; stage_index-- {
		if stages[stage_index] == stage_name {
			return stage_index
		}
	}
	return -1
}
