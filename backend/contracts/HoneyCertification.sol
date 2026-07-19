// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title HoneyCertification
/// @notice Immutable on-chain registry mapping a BeeTrack honey batch id to
/// the SHA256 hashes of its lab PDF and metadata. Only the configured minter
/// (the BeeTrack backend's wallet) may write; anyone may read.
contract HoneyCertification {
    struct Certification {
        bytes32 pdfHash;
        bytes32 metadataHash;
        uint256 timestamp;
        address certifiedBy;
    }

    /// @dev batchID => certification. A zero `timestamp` means "not yet certified" —
    /// this is what `certify` checks to reject duplicate submissions.
    mapping(uint256 => Certification) private certifications;

    /// @dev The only address allowed to call `certify`. Set at deployment, and
    /// rotatable afterward via `setMinter` (e.g. if the backend's wallet/key
    /// is ever replaced) — only by the current minter itself.
    address public minter;

    event CertificationCreated(
        uint256 indexed batchID,
        bytes32 pdfHash,
        bytes32 metadataHash,
        uint256 timestamp,
        address certifiedBy
    );
    event MinterUpdated(address indexed previousMinter, address indexed newMinter);

    error NotMinter();
    error AlreadyCertified(uint256 batchID);
    error NotCertified(uint256 batchID);
    error EmptyHashes();
    error ZeroAddress();

    constructor() {
        minter = msg.sender;
    }

    modifier onlyMinter() {
        if (msg.sender != minter) revert NotMinter();
        _;
    }

    /// @notice Rotates the minter address, e.g. if the backend's wallet/private
    /// key is ever replaced. Only the current minter may call this.
    function setMinter(address newMinter) external onlyMinter {
        if (newMinter == address(0)) revert ZeroAddress();
        emit MinterUpdated(minter, newMinter);
        minter = newMinter;
    }

    /// @notice Permanently records the hashes for batchID. Reverts if batchID
    /// was already certified — this is the contract-level half of the
    /// three-layer idempotency guarantee (see HC-BE-25); it turns an
    /// accidental duplicate submission into a safe no-op-that-reverts rather
    /// than a silent double-certification.
    function certify(uint256 batchID, bytes32 pdfHash, bytes32 metadataHash) external onlyMinter returns (bool) {
        if (pdfHash == bytes32(0) && metadataHash == bytes32(0)) revert EmptyHashes();
        if (certifications[batchID].timestamp != 0) revert AlreadyCertified(batchID);

        certifications[batchID] = Certification({
            pdfHash: pdfHash,
            metadataHash: metadataHash,
            timestamp: block.timestamp,
            certifiedBy: msg.sender
        });

        emit CertificationCreated(batchID, pdfHash, metadataHash, block.timestamp, msg.sender);
        return true;
    }

    /// @notice Returns the stored certification for batchID. Reverts if none exists.
    function getCertification(uint256 batchID)
        external
        view
        returns (bytes32 pdfHash, bytes32 metadataHash, uint256 timestamp, address certifiedBy)
    {
        Certification memory c = certifications[batchID];
        if (c.timestamp == 0) revert NotCertified(batchID);
        return (c.pdfHash, c.metadataHash, c.timestamp, c.certifiedBy);
    }
}
