package thanks

import (
	"context"
	"errors"
	"github.com/google/go-github/github"
)

type org struct {
	mainRepository string
	client *github.Client
	contributors map[string][]*github.ContributorStats
	filterStats map[string][]string
}

type thanks interface {
	ListRepository(org string) ([]*github.Repository, error)
	ListRelease(org string) ([]*github.RepositoryRelease, error)
	ListContributorsStats(org, repo string) (error)
	FilterContributors(currentRelease,oldRelease github.RepositoryRelease, repo string) (error)
	Thanks() map[string][]string
}

func NewReleaseClient(mainRepository string) thanks {
	return org {
		mainRepository: mainRepository,
		client: github.NewClient(nil),
		contributors:  map[string][]*github.ContributorStats{},
		filterStats: map[string][]string{},
	}
}

func (o org) ListRepository(org string) ([]*github.Repository, error){
	opt := &github.RepositoryListByOrgOptions{Type: "public"}
	repos, _, err := o.client.Repositories.ListByOrg(context.Background(), org, opt)
	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return []*github.Repository{}, errors.New("hit rate limit")
		}
		return []*github.Repository{}, err
	}
	return repos, nil
}

func (o org) ListRelease(org string) ([]*github.RepositoryRelease, error){
	opt := &github.ListOptions{Page: 1, PerPage: 100}
	releases,_, err := o.client.Repositories.ListReleases(context.Background(),org, o.mainRepository, opt)
	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return []*github.RepositoryRelease{}, errors.New("hit rate limit")
		}
		return []*github.RepositoryRelease{}, err
	}
	return releases,nil
}

func (o org) ListContributorsStats(org, repo string) (error){
	stats, _, err := o.client.Repositories.ListContributorsStats(context.Background(), org, repo)
	if err != nil {
		if _, ok := err.(*github.AcceptedError); ok {
			return errors.New("scheduled")
		}
		return err
	}
	if _,ok := o.contributors[repo]; !ok {
		o.contributors[repo]= []*github.ContributorStats{}
	}
	o.contributors[repo] = stats
	return nil
}

func (o org) FilterContributors(currentRelease,oldRelease github.RepositoryRelease, repo string) (error){
	for _,stat := range o.contributors[repo] {
		for _,s := range stat.Weeks {
			if currentRelease.CreatedAt.After(s.Week.Time) && oldRelease.CreatedAt.Before(s.Week.Time) {
				if _, ok := o.filterStats[currentRelease.GetTagName()]; !ok {
					o.filterStats[currentRelease.GetTagName()] = []string{}
				}
				o.filterStats[currentRelease.GetTagName()] = append(o.filterStats[currentRelease.GetTagName()], *stat.GetAuthor().Login)
			}
		}
	}
	return  nil
}


func (o org) Thanks() map[string][]string {
	for release,actors := range o.filterStats {
		o.filterStats[release] = removeDuplicates(actors)
	}
	return o.filterStats
}


func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}


func removeDuplicates(strList []string) []string {
	list := []string{}
	for _, item := range strList {
		if contains(list, item) == false {
			list = append(list, item)
		}
	}
	return list
}
