package chord

import (
	"strings"
	"os"
	"errors"
)


// StoreFile: called by THIS node to upload a file into the ring
func (n *Node) StoreFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Extract filename only
	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1]

	// Compute key = hash(filename)
	key := HashKey(filename, n.m)

	// Lookup responsible node
	target := n.Find(key)
	if target == nil {
		return errors.New("StoreFile: lookup failed")
	}

	req := &StoreFileRequest{
		Filename: filename,
		Hash:     key,
		Data:     data,
	}
	rep := &StoreFileReply{}

	ok := call(target.Addr, "Node.StoreFileRPC", req, rep)
	if !ok || !rep.OK {
		return errors.New("StoreFile: RPC failed")
	}

	return nil
}

// Lookup: return node info + file contents
func (n *Node) Lookup(filename string) (owner *RemoteNode, data []byte, err error) {
	key := HashKey(filename, n.m)

	target := n.Find(key)
	if target == nil {
		return nil, nil, errors.New("Lookup: Find() failed")
	}

	req := &GetFileRequest{Hash: key}
	rep := &GetFileReply{}

	ok := call(target.Addr, "Node.GetFileRPC", req, rep)
	if !ok || !rep.OK {
		return target, nil, errors.New("Lookup: file not found")
	}

	return target, rep.Data, nil
}