
export P1=`ps -elf |grep firecracker | grep -v 'grep' |awk {'print $4'}`

pstack $P1

gdb -p $P1 --batch -ex "set pagination off" -ex "thread apply all bt full" -ex "detach" -ex "quit"

perf top -p $P1

