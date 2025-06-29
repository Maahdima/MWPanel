-- name: GetPeer :one
SELECT * FROM wg_peers
WHERE id = ? LIMIT 1;

-- name: ListPeers :many
SELECT * FROM wg_peers
ORDER BY id;

-- name: CreatePeer :one
INSERT INTO wg_peers (
                peer_id,
                disabled,
                comment,
                peer_name,
                public_key,
                interface,
                allowed_address,
                endpoint,
                endpoint_port,
                persistent_keepalive,
                scheduler_id,
                queue_id,
                expire_time,
                traffic_limit,
                download_bandwidth,
                upload_bandwidth
            ) VALUES (
                ?,
                ?,
                ?,
                ?,
                ?,
                ?,
                ?,
                ?,
                ?,
                ?,
                ?,
                ?,
                ?,
                ?,
                ?,
                ?
            )
RETURNING *;