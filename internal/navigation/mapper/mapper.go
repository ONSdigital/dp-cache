package mapper

import (
	coreModel "github.com/ONSdigital/dp-renderer/model"
	topicModel "github.com/ONSdigital/dp-topic-api/models"
)

func MapNavigationContent(navigationContent topicModel.Navigation) []coreModel.NavigationItem {
	var mappedNavigationContent []coreModel.NavigationItem
	if navigationContent.Items != nil {
		for _, rootContent := range *navigationContent.Items {
			var subItems []coreModel.NavigationItem
			if rootContent.SubtopicItems != nil {
				for _, subtopicContent := range *rootContent.SubtopicItems {
					subItems = append(subItems, coreModel.NavigationItem{
						Uri:   subtopicContent.Uri,
						Label: subtopicContent.Label,
					})
				}
			}
			mappedNavigationContent = append(mappedNavigationContent, coreModel.NavigationItem{
				Uri:      rootContent.Uri,
				Label:    rootContent.Label,
				SubItems: subItems,
			})
		}
	}
	return mappedNavigationContent
}
