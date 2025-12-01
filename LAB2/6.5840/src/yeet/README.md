
# build sequentiall 

cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/main
build plugin ../mrapps/wc.go
go run -race mrsequential.go ../mrapps/wc.so pg-*.txt
# check output




#  **1. Build Plugin (wc.so)**

Rebuild plugin every time code in `mr/` changes:

```bash
cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/mrapps
rm  *.so
go build -race -buildmode=plugin wc.go
```

---

#  **2. Run Basic MapReduce (Single Machine)**

## Start coordinator

```bash
cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/main
go run -race mrcoordinator.go pg-*.txt
```

## Start worker(s)

Open another terminal:

```bash
cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/main
go run -race mrworker.go ../mrapps/wc.so  
```

Run as many workers as you want.

## Check output

```bash
ls mr-out-*
cat mr-out-* | sort > summed
```



#  **4. Manual Correctness Verification**

## Build plugin

```bash
cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/mrapps
rm  *.so
go build -race -buildmode=plugin wc.go
```
```

---

## Sequential reference output

```bash
cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/main
go run -race mrsequential.go ../mrapps/wc.so pg-*.txt
cat mr-out-0 | sort > seq-out
```

---

---

## Compare

```bash
diff seq-out summed
```

No output = correct.

---

#  **5. Advanced Distributed Version (Multi-Machine AWS Deployment)**

Below is a full sequence of commands to deploy:

* 1 coordinator machine
* 2 worker machines

---

# **Build files locally**

```bash
cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/mrapps
go build -race -buildmode=plugin wc.go

cd ../main
go build -race -o advcoordinator advcoordinator.go
go build -race -o advworker advworker.go
```

---

# **Upload files to AWS**

## Upload to coordinator

```bash

cd ~/
ssh -i Downloads/Downloads/labsuser.pem ubuntu@COORD_PUBLIC_IP //3.236.170.122

hostname -I //inside aws to get private IP


Move binaries and plugin to aws instances
cd ~/
scp -i Downloads/labsuser.pem \
~/Desktop/Distributed_systems/LAB2/6.5840/src/main/advcoordinator \
~/Desktop/Distributed_systems/LAB2/6.5840/src/mrapps/wc.so \
~/Desktop/Distributed_systems/LAB2/6.5840/src/main/pg-*.txt \
ubuntu@COORD_PUBLIC_IP:~/  ## 3.236.170.122
```

## Upload to workers

```bash
scp -i Downloads/labsuser.pem \
~/Desktop/Distributed_systems/LAB2/6.5840/src/main/advworker \
~/Desktop/Distributed_systems/LAB2/6.5840/src/mrapps/wc.so \
ubuntu@WORKER1_PUBLIC_IP:~/ ## 100.24.209.41


scp -i Downloads/labsuser.pem \
~/Desktop/Distributed_systems/LAB2/6.5840/src/main/advworker \
~/Desktop/Distributed_systems/LAB2/6.5840/src/mrapps/wc.so \
ubuntu@WORKER2_PUBLIC_IP:~/ ## 44.200.74.246

```

---

# **Verify file placement**

### Coordinator instance

```bash
cd ~/
ssh -i Downloads/labsuser.pem ubuntu@COORD_PUBLIC_IP
ls
```

Should contain: `advcoordinator`, `wc.so`, `pg-xxx.txt`.

### Worker instances

```bash
ssh -i Downloads/labsuser.pem ubuntu@WORKER1_PUBLIC_IP
ls

ssh -i Downloads/labsuser.pem ubuntu@WORKER2_PUBLIC_IP
ls
```

Should contain: `advworker`, `wc.so`.

---

# **Start advanced MapReduce cluster**

## Start coordinator

```bash
ssh -i Downloads/labsuser.pem ubuntu@COORD_PUBLIC_IP
./advcoordinator pg-*.txt
```

Leave it running.

## Start worker(s)

Worker 1:

```bash
ssh -i Downloads/labsuser.pem ubuntu@WORKER1_PUBLIC_IP
./advworker wc.so
```

Worker 2:

```bash
ssh -i Downloads/labsuser.pem ubuntu@WORKER2_PUBLIC_IP
./advworker wc.so
```

Workers will:

* Register their private VPC IP
* Store intermediate map outputs locally
* Serve buckets via RPC (GetBucket)
* Shutdown when coordinator exits

---

# **Fetch output from workers**

```bash
scp -i ~/Downloads/labsuser.pem "ubuntu@worker:~/mr-out-*" .
scp -i ~/Downloads/labsuser.pem "ubuntu@worker:~/mr-out-*" .


cat mr-out-* | sort | head
cat mr-out-0 | sort > aws-out
```

Expected: same as sequential output.

