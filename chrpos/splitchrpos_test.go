package chrpos

import "testing"

func TestLimitedChunkChrRange(t *testing.T) {
	parts, err := ChunkChrRange(1000000, "grch37", "1", 0, 24681012)
	if err != nil {
		t.Errorf("%w", err)
	}

	t.Log(parts)
}

func TestUnlimitedChunkChrRange(t *testing.T) {
	parts, err := ChunkChrRange(10000000, "grch37", "1", 0, 0)
	if err != nil {
		t.Errorf("%w", err)
	}

	t.Log(parts)
}
