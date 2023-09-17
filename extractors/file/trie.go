package file

import "github.com/dghubble/trie"

func GetTrie(words []string) *trie.PathTrie {
	trie := trie.NewPathTrie()
	for _, word := range words {
		trie.Put(word, nil)
	}
	return trie
}
