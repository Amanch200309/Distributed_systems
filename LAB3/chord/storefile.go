package chord

import (
	"errors"
	"os"
	"strings"
)

/*
StoreFile uploads a file into the Chord ring.

	Args: 	path (local file path to upload)
	Returns: error if file read or RPC call fails
	Reads file, hashes filename to find responsible node,
	sends file via RPC to that node for storage
*/
func (n *Node) StoreFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1] // Extract filename from path

	key := HashKey(filename, n.m) // Compute key = hash(filename)

	target := n.Find(key) // Lookup responsible node
	if target == nil {
		return errors.New("StoreFile: lookup failed")
	}

	req := &StoreFileRequest{
		Filename: filename,
		Hash:     key,
		Data:     data,
	}
	rep := &StoreFileReply{}

	// Store on primary node via RPC
	ok := call(target.Addr, "Node.StoreFileRPC", req, rep)
	if !ok || !rep.OK {
		return errors.New("StoreFile: RPC failed")
	}

	// Replicate to successor nodes for fault tolerance
	// Get successor list from target node
	//lagra filen på succesorerna också för backup
	succReq := &GetSuccessorListRequest{}
	succRep := &GetSuccessorListReply{}
	if call(target.Addr, "Node.GetSuccessorListRPC", succReq, succRep) {
		// Replicate to r-1 successors (target already has it)
		replicaCount := 0
		maxReplicas := len(succRep.Successors)

		for _, succ := range succRep.Successors {
			if succ != nil && succ.Addr != target.Addr && replicaCount < maxReplicas {
				replicaRep := &StoreFileReply{}
				call(succ.Addr, "Node.StoreFileRPC", req, replicaRep)
				// Don't fail if replica storage fails - we have primary
				replicaCount++
			}
		}
	}

	return nil
}

/*
Lookup finds the node responsible for a file and retrieves its contents.

	Args: 	filename (name of file to lookup)
	Returns: owner (node storing the file),
			 data (file contents),
			 err (error if lookup or retrieval fails)
	Hashes filename to find key, performs lookup to find owner node,
	retrieves file data via RPC
*/
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
