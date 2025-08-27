cd "$(dirname "$0")/.."

docker run --rm -it \
  --privileged \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v "$HOME/.containerlab:/root/.containerlab" \
  -v "$(pwd):/lab" \
  --network host \
  ghcr.io/srl-labs/clab bash -c "
    cd /lab &&
    containerlab deploy --reconfigure -t lab.yml
  "
