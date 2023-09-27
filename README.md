Disk life eater
========
Command line utility for writing random data to disk indefinitely or until interrupted by the user.
The main purpose is to check disk reliability.

Tested on Linux and Windows.


Usage: diskeater [flags]  
- -b uint  
    - Random pattern size, bytes (default 1024).    
- -h  
    - Help  
- -m uint  
    - Maximum disk space to use(not implemented)  
- -p string  
Junk file prefix (default "KILLSSD")  
- -path string  
    - Path for writing junk files (default "/tmp/")  
- -r  
    - Remove junk on exit (default true)  
- -s uint  
    - Junk file size (default 1Gb)  