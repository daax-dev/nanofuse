  ./bin/nanofuse image list --json | jq -r '.images[].digest' | xargs -I{} ./bin/nanofuse image rm {}
    ./bin/nanofuse vm list --json | jq -r '.vms[].id' | xargs -I{} ./bin/nanofuse vm rm -f {}