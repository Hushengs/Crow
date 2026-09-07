package biz

import "testing"

func TestValidateInjectEnqueue(t *testing.T) {
	if err := validateInjectEnqueue(&InjectEnqueue{ResourceType: InjectResourceVideo, ResourceID: 8, VideoID: 8, Action: InjectActionCreate}); err != nil {
		t.Fatalf("video enqueue: %v", err)
	}
	if err := validateInjectEnqueue(&InjectEnqueue{ResourceType: InjectResourceEpisode, ResourceID: 3, VideoID: 8, EpisodeID: 3, Action: InjectActionUpdate}); err != nil {
		t.Fatalf("episode enqueue: %v", err)
	}
	if err := validateInjectEnqueue(&InjectEnqueue{ResourceType: InjectResourceMedia, ResourceID: 9, VideoID: 8, EpisodeID: 3, MediaID: 9, Action: InjectActionDelete}); err != nil {
		t.Fatalf("media enqueue: %v", err)
	}
	if err := validateInjectEnqueue(&InjectEnqueue{ResourceType: InjectResourceVideo, ResourceID: 8, VideoID: 1, Action: InjectActionCreate}); err == nil {
		t.Fatal("want invalid video enqueue")
	}
}
