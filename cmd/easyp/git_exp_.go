package main

import (
	"fmt"
	"log/slog"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func getRepository() *git.Repository {
	repository, err := git.PlainOpen("/Users/vbliznetsov/Projects/Hound/easyp/easyp")
	if err != nil {
		panic(fmt.Errorf("failed to open repository: %s", err))
	}
	return repository
}

func gitExpDiff() {
	repository := getRepository()

	ref, err := repository.Head()
	if err != nil {
		panic(fmt.Errorf("failed to get HEAD: %s", err))
	}
	_ = ref

	commCur, err := repository.CommitObject(ref.Hash())
	if err != nil {
		panic(fmt.Errorf("failed to get commit object: %s", err))
	}
	_ = commCur

	treeCur, err := commCur.Tree()
	if err != nil {
		panic(fmt.Errorf("failed to get commit tree: %s", err))
	}
	_ = treeCur

	treeCur.Files().ForEach(func(f *object.File) error {
		return nil
	})
}
func gitExp() {
	slog.Info("Git exp")
	gitExpDiff()
	return

	repository := getRepository()
	worktree, err := repository.Worktree()
	if err != nil {
		panic(fmt.Errorf("failed to open worktree: %s", err))
	}
	_ = worktree

	// 36c1bd4
	//hashCommitForCheckout := "36c1bd4"
	//hashCommitForCheckout := "0362c4c53df82d409903992a2c085b54c8a3368d"
	//if err := worktree.Checkout(&git.CheckoutOptions{
	//	Hash: plumbing.NewHash(hashCommitForCheckout),
	//}); err != nil {
	//	panic(fmt.Errorf("failed to checkout commit: %s", err))
	//}
	//
	//return

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

	oldTreeHash := "db7c6675f3954fda0ba13eddcbe8171b55d539cc"
	oldTree, err := repository.TreeObject(plumbing.NewHash(oldTreeHash))
	if err != nil {
		panic(fmt.Errorf("failed to list commits: %s", err))
	}
	_ = oldTree
	oldTree.Files().ForEach(func(f *object.File) error {
		return nil
	})

	newTreeHash := "d86a0b1604ce6873339e7df0bf9b77612f253b83"
	newTree, err := repository.TreeObject(plumbing.NewHash(newTreeHash))
	if err != nil {
		panic(fmt.Errorf("failed to list commits: %s", err))
	}
	_ = newTree

	diff, err := newTree.Diff(oldTree)
	if err != nil {
		panic(fmt.Errorf("failed to list commits: %s", err))
	}
	_ = diff

	//repository.TreeObject()

	//repository.TreeObject().
	//	slog.Info("repo", repository)

	mainBranch, err := repository.Branch("main")
	if err != nil {
		panic(fmt.Errorf("failed to list main branch: %s", err))
	}
	_ = mainBranch
}
