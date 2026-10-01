package service

const customDedicatedRelatedGroupsSQL = `SELECT ag.group_id FROM account_groups ag
 JOIN groups related ON related.id=ag.group_id AND related.deleted_at IS NULL
 WHERE ag.account_id=$1`

const customDedicatedAliasAccessSQL = `SELECT (` + customDedicatedAccessSQL + `)
 AND EXISTS (SELECT 1 FROM account_groups ag JOIN groups g ON g.id=ag.group_id
 WHERE ag.account_id=$1 AND ag.group_id=$5 AND g.deleted_at IS NULL AND g.status='active' AND g.is_exclusive=TRUE)
 AND EXISTS (SELECT 1 FROM user_allowed_groups WHERE user_id=$4 AND group_id=$5)`

const customDedicatedCheckSQL = "SELECT " + customDedicatedColumns + ` FROM custom_dedicated_accounts WHERE (group_id=$1 OR account_id=$2 OR account_id=$3 OR account_id=(SELECT parent_account_id FROM accounts WHERE id=$2)
 OR account_id IN (SELECT account_id FROM account_groups WHERE group_id=$1))
 AND (deleted_at IS NULL OR NOT EXISTS (SELECT 1 FROM custom_dedicated_accounts live
 WHERE live.deleted_at IS NULL AND (live.group_id=$1 OR live.account_id IN (SELECT account_id FROM account_groups WHERE group_id=$1))
 AND ($2::bigint=0 OR live.account_id=$2)))`
