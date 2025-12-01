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
cat mr-out-* | sort | head
```

Expected:

```
A 509
ABOUT 2
ACT 8
```

---

# **3. Run Full Test Suite**

```bash
cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/main
bash test-mr.sh
```

Expected:

```
*** PASSED ALL TESTS
```

---

#  **4. Manual Correctness Verification**

## Build plugin

```bash
cd ~/Desktop/Distributed_systems/LAB2/6.5840/src/main
rm -f wc.so mr-out-* seq-out dist-out
go build -buildmode=plugin ../mrapps/wc.go
```

---

## Sequential reference output

```bash
go run mrsequential.go wc.so pg-*.txt
cat mr-out-0 | sort > seq-out
```

---

## Distributed output (basic version)

```bash
rm -f mr-out-*
go run mrcoordinator.go pg-*.txt
# In another terminal:
go run mrworker.go wc.so

cat mr-out-* | sort > dist-out
```

---

## Compare

```bash
diff seq-out dist-out
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
go build -buildmode=plugin wc.go

cd ../main
go build -o advcoordinator advcoordinator.go
go build -o advworker advworker.go
```

---

# **Upload files to AWS**

## Upload to coordinator

```bash
scp -i mykey.pem \
~/Desktop/Distributed_systems/LAB2/6.5840/src/main/advcoordinator \
~/Desktop/Distributed_systems/LAB2/6.5840/src/mrapps/wc.so \
~/Desktop/Distributed_systems/LAB2/6.5840/src/main/pg-*.txt \
ubuntu@COORD_PUBLIC_IP:~/
```

## Upload to workers

```bash
scp -i mykey.pem \
~/Desktop/Distributed_systems/LAB2/6.5840/src/main/advworker \
~/Desktop/Distributed_systems/LAB2/6.5840/src/mrapps/wc.so \
ubuntu@WORKER1_PUBLIC_IP:~/

scp -i mykey.pem \
~/Desktop/Distributed_systems/LAB2/6.5840/src/main/advworker \
~/Desktop/Distributed_systems/LAB2/6.5840/src/mrapps/wc.so \
ubuntu@WORKER2_PUBLIC_IP:~/
```

---

# **Verify file placement**

### Coordinator instance

```bash
ssh -i mykey.pem ubuntu@COORD_PUBLIC_IP
ls
```

Should contain: `advcoordinator`, `wc.so`, `pg-xxx.txt`.

### Worker instances

```bash
ssh -i mykey.pem ubuntu@WORKER1_PUBLIC_IP
ls

ssh -i mykey.pem ubuntu@WORKER2_PUBLIC_IP
ls
```

Should contain: `advworker`, `wc.so`.

---

# **Start advanced MapReduce cluster**

## Start coordinator

```bash
ssh -i mykey.pem ubuntu@COORD_PUBLIC_IP
./advcoordinator pg-*.txt
```

Leave it running.

## Start worker(s)

Worker 1:

```bash
ssh -i mykey.pem ubuntu@WORKER1_PUBLIC_IP
./advworker wc.so
```

Worker 2:

```bash
ssh -i mykey.pem ubuntu@WORKER2_PUBLIC_IP
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
scp -i mykey.pem ubuntu@COORD_PUBLIC_IP:~/mr-out-* .
cat mr-out-* | sort | head
```

Expected: same as sequential output.

