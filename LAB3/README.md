## create a node   
./main -a 127.0.0.1 -p 4170 --ts 3000 --tff 1000 --tcp 3000 -r 4

## join a node
./main -a 127.0.0.1 -p 4171 --ja 127.0.0.1 --jp 4170 --ts 3000 --tff 1000 --tcp 3000 -r 4


./main -a 127.0.0.1 -p 4172 --ja 127.0.0.1 --jp 4170 --ts 3000 --tff 1000 --tcp 3000 -r 4