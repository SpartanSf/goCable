package player

import (
	"errors"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"sync"
	"fmt"
)

type MediaListSortStrategy interface {
	Sort([]string)
}

type SortStratRandom struct{}

func (s SortStratRandom) Sort(list []string) {
	rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
}

type MediaList struct {
	list         []string
	nextList     []string
	current      int
	SortStrategy MediaListSortStrategy

	mu sync.Mutex
}

func NewMediaList(list []string, sortStrat MediaListSortStrategy) (*MediaList, error) {
	if len(list) == 0 {
		return nil, errors.New("need media")
	}
	ml := &MediaList{
		list:         append([]string(nil), list...),
		nextList:     append([]string(nil), list...),
		SortStrategy: sortStrat,
	}
	ml.SortStrategy.Sort(ml.list)
	ml.SortStrategy.Sort(ml.nextList)
	return ml, nil
}


func (ml *MediaList) All() []string {
	return ml.list
}

func (ml *MediaList) Current() string {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	return ml.list[ml.current]
}

func (ml *MediaList) Next() string {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if ml.current+1 >= len(ml.list) {
		if len(ml.nextList) == 0 {
			fmt.Println("[debug] Next() called, but nextList is empty")
			return ""
		}
		return ml.nextList[0]
	}
	return ml.list[ml.current+1]
}


func (ml *MediaList) Advance() string {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if ml.current+1 >= len(ml.list) {
		if len(ml.nextList) == 0 {
			fmt.Println("[debug] MediaList.Advance: nextList is empty, staying at current index")
			ml.current = 0
			if len(ml.list) == 0 {
				return ""
			}
			return ml.list[ml.current]
		}

		ml.list, ml.nextList = ml.nextList, ml.list
		ml.SortStrategy.Sort(ml.nextList)
		ml.current = 0
	} else {
		ml.current++
	}

	if ml.current >= len(ml.list) {
		fmt.Println("[debug] MediaList.Advance: current index out of bounds after increment")
		return ""
	}

	return ml.list[ml.current]
}


var VideoFiles map[string]struct{} = map[string]struct{}{
	".avi": {},
	".mp4": {},
	".mkv": {},
}

func FromFolder(folderPath string) (*MediaList, error) {
	var paths []string
	if err := filepath.Walk(folderPath, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if _, ok := VideoFiles[path.Ext(file)]; ok {
			paths = append(paths, file)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return NewMediaList(paths, SortStratRandom{})
}
