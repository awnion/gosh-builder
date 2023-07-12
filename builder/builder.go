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

func Build(repo_url string) (target_sha string, err error) {
	// TODO: parse repo_url with #<tree-hash>:<path/dir>
	if !strings.HasPrefix(repo_url, "gosh://") {
		return "", errors.New(fmt.Sprint("repo_url must be gosh://", repo_url))
	}

	// create a new directory for the repo
	image_cache_dir := getCacheDir(repo_url)
	log.Println("image_cache_dir", image_cache_dir)

	if err := os.MkdirAll(image_cache_dir, 0777); err != nil {
		return "", err
	}

	repo_dir := image_cache_dir + "/repo"
	if _, err := os.Stat(repo_dir); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", repo_url, repo_dir)
		cmd.Dir = image_cache_dir
		if err := cmd.Run(); err != nil {
			return "", err
		}
	} else {
		cmd := exec.Command("git", "pull")
		cmd.Dir = repo_dir
		if err := cmd.Run(); err != nil {
			return "", err
		}
	}

	if err := os.Chdir(repo_dir); err != nil {
		return "", err
	}

	dockerfile_path := repo_dir + "/Dockerfile"
	df, err := os.ReadFile(dockerfile_path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			msg := fmt.Sprint("Dockerfile not found in", repo_dir)
			return "", errors.New(msg)
		} else {
			return "", err
		}
	}

	sbom_path := repo_dir + "/sbom.spdx.json"
	sbom_content, err := os.ReadFile(sbom_path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			msg := fmt.Sprint("SBOM not found in", repo_dir)
			return "", errors.New(msg)
		} else {
			return "", err
		}
	}
	sbom, err := sbom.ParseSBOM(sbom_content)
	if err != nil {
		return "", err
	}

	// TODO: !!!commit precision required
	image_lock_mapping := make(map[string]string)
	for _, component := range sbom.Components {
		if component.Type == "image" {
			image_lock_mapping[component.Name] = component.Purl
		}
	}
	log.Println("image_lock_mapping", image_lock_mapping)

	image_docker_sha_mapping := make(map[string]string)

	dockerfile, err := parser.Parse(bytes.NewReader([]byte(df)))
	if err != nil {
		return "", err
	}

	// TODO: implement stage aliasing
	for _, child := range dockerfile.AST.Children {
		if strings.ToUpper(child.Value) == "FROM" {
			child.Dump()
		}
	}

	// parse Dockerfile
	// replace all FROM gosh to Build() result sha

	// make a new docker_file
	docker_file_lock := ""
	dockerfile_lock_path := image_cache_dir + "/lock.Dockerfile"
	if err := os.WriteFile(dockerfile_lock_path, []byte(docker_file_lock), 0777); err != nil {
		return "", err
	}

	target_sha, err = dockerLockBuild([]byte(dockerfile_lock_path), repo_dir)
	// docker build with new docker file

	target_sha_path := image_cache_dir + "/.target_sha"
	log.Println("Write docker image hash to", target_sha_path)
	if err = os.WriteFile(target_sha_path, []byte(target_sha), 0777); err != nil {
		return "", err
	}

	return target_sha, nil
}

func dockerLockBuild(dockerfile_lock_path, repo_dir string) (target_sha string, err error) {
	// TODO: exec.CommandContext()
	cmd := exec.Command("docker", "buildx", "build", "-f", dockerfile_lock_path, repo_dir)
	cmd.Dir = repo_dir
	cmd.Stdin = bytes.NewReader(dockerfile_lock)

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
