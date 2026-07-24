-- Wipes all dev data except honey batches that already have a confirmed
-- on-chain certification (and whatever rows those batches need to stay
-- valid: their owning user, certification history, blockchain jobs, QR
-- codes, and certification requests). Confirmed batches are permanent
-- on-chain records — losing them locally on every reset means the next
-- certify() for a batch that reuses that same id reverts as
-- "already certified" against unrelated content (see writer.go /
-- blockchain_worker.go's hash-mismatch safety net). Leaving them in place
-- also means honey_batches' id sequence is never rewound, so a fresh batch
-- can never collide with an old certified id again.
--
-- Run via: docker exec -i <db-container> psql -U postgres -d beetrack -f -
-- (see reset-dev-db.ps1). goose_db_version is never touched.

DELETE FROM inspection_diseases;
DELETE FROM inspection_images;
DELETE FROM hive_diseases;
DELETE FROM inspections;
DELETE FROM treatments;
DELETE FROM feedings;
DELETE FROM harvests;
DELETE FROM apiary_invitations;
DELETE FROM apiary_members;
DELETE FROM listing_favorites;
DELETE FROM listing_images;
DELETE FROM listings;
DELETE FROM hives;
DELETE FROM apiaries;

-- Cascades honey_batch_certifications, honey_batch_qr_codes,
-- honey_batch_certification_requests, and blockchain_jobs for every batch
-- deleted here — only batches with a confirmed certification survive.
DELETE FROM honey_batches
WHERE id NOT IN (
    SELECT DISTINCT batch_id FROM honey_batch_certifications WHERE status = 'confirmed'
);

-- Soft-delete the survivors so they stay out of the reseeded user's honey
-- batch list (every repository read already filters `deleted_at IS NULL`)
-- without touching their row or id — the blockchain worker's id-collision
-- check reads via GetByIDIgnoringDeletion, which is unaffected by this.
UPDATE honey_batches SET deleted_at = NOW() WHERE deleted_at IS NULL;

-- Keep every user still referenced by a surviving honey batch or its
-- certification request history; everyone else can be wiped and reseeded.
DELETE FROM users
WHERE id NOT IN (
    SELECT user_id FROM honey_batches
    UNION
    SELECT requested_by FROM honey_batch_certification_requests WHERE requested_by IS NOT NULL
    UNION
    SELECT reviewed_by FROM honey_batch_certification_requests WHERE reviewed_by IS NOT NULL
);
