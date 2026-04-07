-- 003: OAuth (Google) + Team invites
-- Makes org_id nullable so users exist before joining a team.
-- Adds Google OAuth columns. Creates team_invites table.

-- 1. Users: OAuth fields
ALTER TABLE users ADD COLUMN IF NOT EXISTS google_id TEXT UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT NOT NULL DEFAULT '';

-- 2. Users: make org_id nullable (user exists before joining a team)
ALTER TABLE users ALTER COLUMN org_id DROP NOT NULL;

-- 3. Users: make password_hash optional (OAuth users have no password)
ALTER TABLE users ALTER COLUMN password_hash SET DEFAULT '';

-- 4. Team invite codes
CREATE TABLE IF NOT EXISTS team_invites (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  code TEXT NOT NULL UNIQUE,
  created_by UUID NOT NULL REFERENCES users(id),
  expires_at TIMESTAMPTZ NOT NULL,
  max_uses INTEGER NOT NULL DEFAULT 0,
  use_count INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_team_invites_code ON team_invites (code);
CREATE INDEX IF NOT EXISTS idx_team_invites_org_id ON team_invites (org_id);
