
function for-each-peer() {
    for i in $(ls -d --color=never *); do
        if [[ -d $i ]]; then
            echo ip netns exec $i $@
        fi
    done
}
