// Copyright (c) 2012-2013 Jason McVetta.  This is Free Software, released under
// the terms of the GPL v3.  See http://www.gnu.org/copyleft/gpl.html for details.

package neo4j

import (
	"github.com/jmcvetta/restclient"
	"net/url"
)

func (db *Database) CreateNodeIndex(name, idxType, provider string) (*NodeIndex, error) {
	idx, err := db.createIndex(db.HrefNodeIndex, name, idxType, provider)
	if err != nil {
		return nil, err
	}
	return &NodeIndex{*idx}, nil
}

func (db *Database) NodeIndexes() ([]*NodeIndex, error) {
	indexes, err := db.indexes(db.HrefNodeIndex)
	if err != nil {
		return nil, err
	}
	nis := make([]*NodeIndex, len(indexes))
	for i, idx := range indexes {
		nis[i] = &NodeIndex{*idx}
	}
	return nis, nil
}

func (db *Database) NodeIndex(name string) (*NodeIndex, error) {
	idx, err := db.index(db.HrefNodeIndex, name)
	if err != nil {
		return nil, err
	}
	return &NodeIndex{*idx}, nil

}

// A NodeIndex is an index for searching Nodes.
type NodeIndex struct {
	index
}

// Add associates a Node with the given key/value pair in the given index.
func (nix *NodeIndex) Add(n *Node, key, value string) error {
	return nix.add(n, key, value)
}

// Remove deletes all entries with a given node, key and value from the index.
// If value or both key and value are the blank string, they are ignored.
func (nix *NodeIndex) Remove(n *Node, key, value string) error {
	return nix.remove(n, key, value)
}

// Find locates Nodes in the index by exact key/value match.
func (idx *NodeIndex) Find(key, value string) (map[int]*Node, error) {
	nm := make(map[int]*Node)
	rawurl, err := idx.uri()
	if err != nil {
		return nm, err
	}
	rawurl = join(rawurl, key, value)
	u, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return nm, err
	}
	ne := new(neoError)
	resp := []nodeResponse{}
	req := restclient.RequestResponse{
		Url:    u.String(),
		Method: "GET",
		Result: &resp,
		Error:  ne,
	}
	status, err := idx.db.rc.Do(&req)
	if err != nil {
		logPretty(ne)
		return nm, err
	}
	if status != 200 {
		logPretty(req)
		return nm, BadResponse
	}
	for _, r := range resp {
		n := Node{}
		n.db = idx.db
		n.populate(&r)
		nm[n.Id()] = &n
	}
	return nm, nil
}

// Query locatess Nodes by query, in the query language appropriate for a given Index.
func (idx *index) Query(query string) (map[int]*Node, error) {
	nm := make(map[int]*Node)
	rawurl, err := idx.uri()
	if err != nil {
		return nm, err
	}
	v := make(url.Values)
	v.Add("query", query)
	rawurl += "?" + v.Encode()
	u, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return nm, err
	}
	result := []nodeResponse{}
	req := restclient.RequestResponse{
		Url:    u.String(),
		Method: "GET",
		Result: &result,
	}
	status, err := idx.db.rc.Do(&req)
	if err != nil {
		return nm, err
	}
	if status != 200 {
		logPretty(req)
		return nm, BadResponse
	}
	for _, r := range result {
		n := Node{}
		n.db = idx.db
		n.populate(&r)
		nm[n.Id()] = &n
	}
	return nm, nil
}