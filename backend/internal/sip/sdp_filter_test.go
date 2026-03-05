package sip

import (
	"strings"
	"testing"
)

func TestFilterSDPCandidates_NoLimit(t *testing.T) {
	sdp := "v=0\r\nm=audio 9 UDP/TLS/RTP/SAVPF 111\r\na=candidate:1 1 UDP 2130706431 192.0.2.1 54321 typ host\r\na=candidate:2 1 UDP 1677724415 203.0.113.1 54321 typ srflx\r\na=rtcp-mux\r\n"
	got := filterSDPCandidates(sdp, 0)
	if got != sdp {
		t.Errorf("expected SDP unchanged with limit=0, got:\n%s", got)
	}
}

func TestFilterSDPCandidates_LimitKeepsFirstNCandidates(t *testing.T) {
	sdp := strings.Join([]string{
		"v=0",
		"m=audio 9 UDP/TLS/RTP/SAVPF 111",
		"a=candidate:1 1 UDP 2130706431 192.0.2.1 54321 typ host",
		"a=candidate:2 1 UDP 1677724415 203.0.113.1 54321 typ srflx",
		"a=candidate:3 1 UDP 16777215 10.0.0.1 54321 typ relay",
		"a=rtcp-mux",
	}, "\r\n")

	got := filterSDPCandidates(sdp, 2)

	candidateCount := 0
	for _, line := range strings.Split(got, "\r\n") {
		if strings.HasPrefix(line, "a=candidate:") {
			candidateCount++
		}
	}
	if candidateCount != 2 {
		t.Errorf("expected 2 candidates after filtering, got %d\nSDP:\n%s", candidateCount, got)
	}
	// The first two candidates (highest priority) must be kept.
	if !strings.Contains(got, "a=candidate:1") {
		t.Error("expected first candidate to be kept")
	}
	if !strings.Contains(got, "a=candidate:2") {
		t.Error("expected second candidate to be kept")
	}
	if strings.Contains(got, "a=candidate:3") {
		t.Error("expected third candidate to be dropped")
	}
	// Non-candidate lines must be preserved.
	if !strings.Contains(got, "a=rtcp-mux") {
		t.Error("expected non-candidate line a=rtcp-mux to be preserved")
	}
}

func TestFilterSDPCandidates_MultipleMediaSections(t *testing.T) {
	sdp := strings.Join([]string{
		"v=0",
		"m=audio 9 UDP/TLS/RTP/SAVPF 111",
		"a=candidate:1 1 UDP 2130706431 192.0.2.1 5000 typ host",
		"a=candidate:2 1 UDP 1677724415 203.0.113.1 5000 typ srflx",
		"a=candidate:3 1 UDP 16777215 10.0.0.1 5000 typ relay",
		"m=video 9 UDP/TLS/RTP/SAVPF 96",
		"a=candidate:4 1 UDP 2130706431 192.0.2.1 5001 typ host",
		"a=candidate:5 1 UDP 1677724415 203.0.113.1 5001 typ srflx",
		"a=candidate:6 1 UDP 16777215 10.0.0.1 5001 typ relay",
	}, "\r\n")

	got := filterSDPCandidates(sdp, 1)

	var candidates []string
	for _, line := range strings.Split(got, "\r\n") {
		if strings.HasPrefix(line, "a=candidate:") {
			candidates = append(candidates, line)
		}
	}
	// 1 candidate per media section × 2 sections = 2 total.
	if len(candidates) != 2 {
		t.Errorf("expected 2 candidates total (1 per section), got %d: %v", len(candidates), candidates)
	}
}

func TestFilterSDPCandidates_PreservesLineEndings_CRLF(t *testing.T) {
	sdp := "v=0\r\nm=audio 9 UDP 111\r\na=candidate:1 1 UDP 100 1.2.3.4 5000 typ host\r\na=candidate:2 1 UDP 50 1.2.3.4 5001 typ srflx\r\n"
	got := filterSDPCandidates(sdp, 1)
	if !strings.Contains(got, "\r\n") {
		t.Error("expected CRLF line endings to be preserved")
	}
}

func TestFilterSDPCandidates_PreservesLineEndings_LF(t *testing.T) {
	sdp := "v=0\nm=audio 9 UDP 111\na=candidate:1 1 UDP 100 1.2.3.4 5000 typ host\na=candidate:2 1 UDP 50 1.2.3.4 5001 typ srflx\n"
	got := filterSDPCandidates(sdp, 1)
	if strings.Contains(got, "\r\n") {
		t.Error("expected LF-only line endings to be preserved (no CRLF)")
	}
}

func TestFilterSDPCandidates_LimitHigherThanCandidateCount(t *testing.T) {
	sdp := "v=0\r\nm=audio 9 UDP 111\r\na=candidate:1 1 UDP 100 1.2.3.4 5000 typ host\r\n"
	got := filterSDPCandidates(sdp, 10)
	// Nothing should be dropped when limit > actual count.
	if got != sdp {
		t.Errorf("expected SDP unchanged when limit exceeds candidate count\ngot:\n%s", got)
	}
}
