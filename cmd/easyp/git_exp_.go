package main

import (
	"fmt"
	"log/slog"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func gitExp() {
	slog.Info("Git exp")

	repository, err := git.PlainOpen("/Users/vbliznetsov/Projects/Hound/easyp/easyp")
	if err != nil {
		panic(fmt.Errorf("failed to open repository: %s", err))
	}

	worktree, err := repository.Worktree()
	if err != nil {
		panic(fmt.Errorf("failed to open worktree: %s", err))
	}

	// 36c1bd4
	//hashCommitForCheckout := "36c1bd4"
	hashCommitForCheckout := "0362c4c53df82d409903992a2c085b54c8a3368d"
	if err := worktree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(hashCommitForCheckout),
	}); err != nil {
		panic(fmt.Errorf("failed to checkout commit: %s", err))
	}

	return

	var hashCommit plumbing.Hash
	cnt := 0

	commits, err := repository.CommitObjects()
	if err != nil {
		panic(fmt.Errorf("failed to list commits: %s", err))
	}
	if err := commits.ForEach(func(ref *object.Commit) error {
		if cnt > 0 {
			return nil
		}

		hashCommit = ref.TreeHash
		cnt++
		_ = ref
		return nil
	}); err != nil {
		panic(fmt.Errorf("failed to list commits: %s", err))
	}

	tree, err := repository.TreeObject(hashCommit)
	if err != nil {
		panic(fmt.Errorf("failed to list commits: %s", err))
	}
	tree.Files().ForEach(func(f *object.File) error {
		return nil
	})

	//repository.TreeObject().
	//	slog.Info("repo", repository)
}
