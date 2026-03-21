
export P1=`ps -elf |grep firecracker | grep -v 'grep' |awk {'print $4'}`
export P2=`ps -elf |grep firecracker | grep -v 'grep' |awk {'print $5'}`

echo " process $P1 - $P2"


for PID in "$P1" "$P2"; do
    echo "[1] # readlink -f /proc/$PID/exe"
    readlink -f /proc/$PID/exe        # binary path
    echo "[2] # cat /proc/$PID/cmdline | tr '\0' ' ' ; echo" 
    cat /proc/$PID/cmdline | tr '\0' ' ' ; echo
    echo "[3] # cat /proc/$PID/environ | tr '\0' '\n'  # environment "
    cat /proc/$PID/environ | tr '\0' '\n'  # environment
    echo "[4] # ls -l /proc/$PID/fd               # open file descriptors "
    ls -l /proc/$PID/fd               # open file descriptors
    echo "[5] # ls -l /proc/$PID/fdinfo           # fd meta (offsets, flags)"
    ls -l /proc/$PID/fdinfo           # fd meta (offsets, flags)
    echo "[6] # cat /proc/$PID/status             # status / uids / threads "
    cat /proc/$PID/status             # status / uids / threads
    echo "[7] # cat /proc/$PID/limits             # resource limits "
    cat /proc/$PID/limits             # resource limits
    echo "[8] # cat /proc/$PID/mountinfo          # mountns info"
    cat /proc/$PID/mountinfo          # mountns info



    lsof -p $PID                 # human readable open files + network
    ss -xp | grep $PID -B2 -A2 
    pmap -x $PID                      # memory map summary
    cat /proc/$PID/smaps              # detailed per-VMA memory (RSS, anon, file...)
    grep VmRSS /proc/$PID/status  
  #  strace -ff -s 200 -p $PID -o /tmp/strace.$PID.log
  #  ltrace -p $PID
done