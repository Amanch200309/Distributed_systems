# Recreating the EC2 Server + Proxy Test Setup

This guide explains how to recreate the distributed systems test environment with:

- Server EC2 instance running a file server (Docker container)
- Proxy EC2 instance forwarding requests (Docker container)
- Testing with curl -x to verify proxy forwarding

Both Docker images are assumed to be stored as .tar files locally.

---

## 1. Launch Two EC2 Instances

Create two Amazon Linux 2 EC2 instances:

### Instance A – Server
Used to serve files  
Port: 8080

### Instance B – Proxy
Used to forward requests  
Port: 8081

When launching each instance:
- Use or create a key pair (labsuser.pem)
- Allow SSH (port 22)

---

## 2. Configure Security Groups

Add inbound rules:

### Server (Instance A)
Custom TCP, port 8080, source 0.0.0.0/0

### Proxy (Instance B)
Custom TCP, port 8081, source 0.0.0.0/0

Both:
SSH, port 22, source 0.0.0.0/0

---

## 3. SSH Into Each Instance

From your computer:

ssh -i labsuser.pem ec2-user@<public-ip>

Replace <public-ip> with the instance’s IPv4 address.

---

## 4. Install Docker (on both instances)

Run on each EC2:

sudo yum update -y
sudo yum install docker -y
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -aG docker ec2-user
exit

Reconnect:

ssh -i labsuser.pem ec2-user@<public-ip>

---

## 5. Upload Your Docker Image Tar Files

From your local machine:

scp -i labsuser.pem server_image.tar ec2-user@<server-ip>:/home/ec2-user/
scp -i labsuser.pem proxy_image.tar ec2-user@<proxy-ip>:/home/ec2-user/

---

## 6. Load the Docker Images

### On the Server instance:

docker load -i server_image.tar

### On the Proxy instance:

docker load -i proxy_image.tar

Note the image names printed after loading (e.g., myserver:latest).

---

## 7. Start the Containers

### On the Server instance (port 8080):

docker run -d -p 8080:8080 myserver:latest

### On the Proxy instance (port 8081):

docker run -d -p 8081:8081 myproxy:latest

---

## 8. Test Server Locally

On the server EC2:

curl http://localhost:8080/test.txt

Should return the file content.

---

## 9. Test Proxy Functionality

From your local machine:

curl -X GET http://<server-ip>:8080/test.txt -x <proxy-ip>:8081

Example:

curl -X GET http://100.27.249.141:8080/test.txt -x 44.197.231.98:8081

If this returns the file, the proxy is working correctly.

---

## 10. Optional: Check Logs

Proxy logs:

docker logs <proxy-contain
