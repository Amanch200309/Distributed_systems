package chord

func findSuccessor(id string) {
	// TODO
}

func closestPrecedingNode(id string) {
	// TODO
}

//pseufdo code iterative for findSuccessor

/*
   // ask node n to find the successor of id
    // or a better node to continue the search with
    n.find_successor(id)
        if (id ∈ (n, successor])
            return true, successor;
        else
            return false, closest_preceding_node(id);

    // search the local table for the highest predecessor of id
    n.closest_preceding_node(id)
        // skip this loop if you do not have finger tables implemented yet
        for i = m downto 1
            if (finger[i] ∈ (n,id))
                return finger[i];
        return successor;

    // find the successor of id
    find(id, start)
        found, nextNode = false, start;
        i = 0
        while not found and i < maxSteps
            found, nextNode = nextNode.find_successor(id);
            i += 1
        if found
            return nextNode;
        else
            report error;
*/
