package thanks

import (
	"context"
	"errors"
	"github.com/google/go-github/github"
)

type org struct {
	owner string
	releaseRepo string
	client *github.Client
	contributors map[string][]*github.ContributorStats
	filterStats map[string]contribution
}

type contribution struct {
	contributors map[string]int64
}

type thanks interface {
	listRepository() ([]*github.Repository, error)
	listRelease(prerelease bool) ([]*github.RepositoryRelease, error)
	listContributorsStats(repo string) error
	filterContributors(currentRelease,oldRelease *github.RepositoryRelease, repo string) error
	Thanks(prerelease bool) (map[string]contribution, error)
}

func NewReleaseClient(owner,releaseRepo string) thanks {
	return org {
		owner: owner,
		releaseRepo: releaseRepo,
		client: github.NewClient(nil),
		contributors:  map[string][]*github.ContributorStats{},
		filterStats: map[string]contribution{},
	}
}

func (o org) listRepository() ([]*github.Repository, error){
	opt := &github.RepositoryListByOrgOptions{Type: "public"}
	repos, _, err := o.client.Repositories.ListByOrg(context.Background(), o.owner, opt)
	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return []*github.Repository{}, errors.New("please try again, You hit github rate limit")
		}
		return []*github.Repository{}, err
	}
	return repos, nil
}

func (o org) listRelease(prerelease bool) ([]*github.RepositoryRelease, error){
	opt := &github.ListOptions{Page: 1, PerPage: 200}
	releases,_, err := o.client.Repositories.ListReleases(context.Background(),o.owner, o.releaseRepo, opt)
	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			return []*github.RepositoryRelease{}, errors.New("please try again, You hit github rate limit")
		}
		return []*github.RepositoryRelease{}, err
	}
	if prerelease {
		filterReleases := []*github.RepositoryRelease{}
		for _,release := range releases {
			if !release.GetPrerelease()  {
				filterReleases = append(filterReleases, release)
			}
		}
		return filterReleases,nil
	}

	return releases,nil
}

func (o org) listContributorsStats(repo string) error{
	stats, _, err := o.client.Repositories.ListContributorsStats(context.Background(), o.owner, repo)
	if err != nil {
		if _, ok := err.(*github.AcceptedError); ok {
			return errors.New("please try in a while, Github scheduled the data collection")
		}
		return err
	}
	if _,ok := o.contributors[repo]; !ok {
		o.contributors[repo]= []*github.ContributorStats{}
	}
	o.contributors[repo] = stats
	return nil
}

func (o org) filterContributors(currentRelease, oldRelease *github.RepositoryRelease, repo string) error {
	var total int64 = 0
	for _,stat := range o.contributors[repo] {
		for _,s := range stat.Weeks {
			if currentRelease.CreatedAt.After(s.Week.Time) && oldRelease.CreatedAt.Before(s.Week.Time) {
				if _, ok := o.filterStats[currentRelease.GetTagName()]; !ok {
					o.filterStats[currentRelease.GetTagName()] = contribution{
						contributors: map[string]int64{},
					}
				}
				if _, ok := o.filterStats[currentRelease.GetTagName()].contributors[stat.GetAuthor().GetLogin()]; !ok {
					o.filterStats[currentRelease.GetTagName()].contributors[stat.GetAuthor().GetLogin()] = 0
				}
				total++
				o.filterStats[currentRelease.GetTagName()].contributors[stat.GetAuthor().GetLogin()] = o.filterStats[currentRelease.GetTagName()].contributors[stat.GetAuthor().GetLogin()] + 1

			}
		}
	}
	return  nil
}

func (o org) Thanks(prerelease bool) (map[string]contribution, error) {
	repositories ,err := o.listRepository()
	if err != nil {
		return o.filterStats,err
	}
	releases ,err := o.listRelease(prerelease)
	if err != nil {
		return o.filterStats,err
	}

	for _, repository := range repositories {
		if err := o.listContributorsStats(repository.GetName()); err != nil {
			return o.filterStats,err
		}
		for i := 0; i < len(releases)-1; i++ {
			if err := o.filterContributors(releases[i], releases[i+1], repository.GetName()); err != nil {
				return o.filterStats,err
			}
		}
	}
	return o.filterStats,nil
}
