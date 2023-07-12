# ~/.cache/gosh/builder/uuid4 -> gosh clone -> gosh build -> docker tag@sha

# rec_build () {
#         mkdir -p ~/.cache/gosh/builder/(repo_url_hash)/.target_sha
#         mkdir -p ~/.cache/gosh/builder/(repo_url_hash)/repo/

#         cd
#
#         for all {
#                 sbom_hash_url := get_sbom_hash_url(origin_url)
#                 docker_image_sha := rec_build(sbom_hash_url)
#                 db.append((origin_url, docker_image_sha))
#         }
#         new_docker_file := replace_all(db) # origin_url -> docker_img_sha
#         bash -c docker build -f ~/.cache/gosh/builder...new_docker_file .
#         return docker_image_sha
# }

# origin_url -> gosh_commit_url... commit_hash.sbom
# tagname@sha:...
fRoM gosh://0:b00a7a5a24740e4a7d6487d31969732f1febcaea412df5cc307400818055ad58/awnion/telepresence-build-gosh#dev:os/ubuntu \
        as builder

COPY . .
RUN --mount=type=bind,targ cargo build --offline

FROM gosh://0:b00a7a5a24740e4a7d6487d31969732f1febcaea412df5cc307400818055ad58/awnion/telepresence-build-gosh#dev:os/ubuntu \
        as final

COPY --from=builder / /

FROM scratch
COPY --from=final / /
