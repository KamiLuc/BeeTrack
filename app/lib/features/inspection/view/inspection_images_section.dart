import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';

import '../../../core/validation/size_tiers.dart';
import '../../../core/widgets/photo_size_snackbar.dart';
import '../../../l10n/app_localizations.dart';
import '../data/inspection_image_model.dart';
import '../data/inspection_image_repository.dart';
import '../data/inspection_model.dart';

Future<void> showInspectionPhotosSheet(
  BuildContext context, {
  required int apiaryId,
  required int hiveId,
  required Inspection inspection,
  required InspectionImageRepository imageRepo,
}) {
  return showModalBottomSheet(
    context: context,
    builder: (_) => _InspectionPhotosSheet(
      apiaryId: apiaryId,
      hiveId: hiveId,
      inspection: inspection,
      imageRepo: imageRepo,
    ),
  );
}

class _InspectionPhotosSheet extends StatefulWidget {
  final int apiaryId;
  final int hiveId;
  final Inspection inspection;
  final InspectionImageRepository imageRepo;

  const _InspectionPhotosSheet({
    required this.apiaryId,
    required this.hiveId,
    required this.inspection,
    required this.imageRepo,
  });

  @override
  State<_InspectionPhotosSheet> createState() => _InspectionPhotosSheetState();
}

class _InspectionPhotosSheetState extends State<_InspectionPhotosSheet> {
  late final Future<List<InspectionImage>> _future;

  @override
  void initState() {
    super.initState();
    _future = widget.imageRepo.listImages(
      widget.apiaryId,
      widget.hiveId,
      widget.inspection.id,
    );
  }

  void _openViewer(List<InspectionImage> images, int index) {
    Navigator.of(context).push(
      PageRouteBuilder(
        opaque: false,
        barrierColor: Colors.black87,
        pageBuilder: (_, __, ___) => InspectionPhotoViewer(
          existingImages: images,
          pendingFiles: const [],
          initialIndex: index,
          urlBuilder: (img) => widget.imageRepo.imageUrl(
            widget.apiaryId,
            widget.hiveId,
            widget.inspection.id,
            img.id,
          ),
          headers: widget.imageRepo.authHeaders(),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;

    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              l10n.inspectionPhotos,
              style: textTheme.titleSmall?.copyWith(color: colorScheme.primary),
            ),
            const SizedBox(height: 12),
            FutureBuilder<List<InspectionImage>>(
              future: _future,
              builder: (context, snapshot) {
                if (snapshot.connectionState != ConnectionState.done) {
                  return const Padding(
                    padding: EdgeInsets.symmetric(vertical: 24),
                    child: Center(child: CircularProgressIndicator()),
                  );
                }
                final images = snapshot.data ?? [];
                if (images.isEmpty) {
                  return Padding(
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    child: Center(
                      child: Text(
                        l10n.inspectionNoPhotos,
                        style: textTheme.bodyMedium?.copyWith(
                          color: colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ),
                  );
                }
                final headers = widget.imageRepo.authHeaders();
                return SizedBox(
                  height: 100,
                  child: ListView.builder(
                    scrollDirection: Axis.horizontal,
                    itemCount: images.length,
                    itemBuilder: (_, i) => Padding(
                      padding: const EdgeInsets.only(right: 8),
                      child: GestureDetector(
                        onTap: () => _openViewer(images, i),
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(8),
                          child: Image.network(
                            widget.imageRepo.imageUrl(
                              widget.apiaryId,
                              widget.hiveId,
                              widget.inspection.id,
                              images[i].id,
                            ),
                            headers: headers,
                            width: 90,
                            height: 90,
                            fit: BoxFit.cover,
                            errorBuilder: (_, __, ___) =>
                                const Icon(Icons.broken_image_outlined),
                          ),
                        ),
                      ),
                    ),
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

const int _kMaxPhotos = 6;

/// Horizontal thumbnail strip with add/delete used inside the inspection form.
class InspectionImagesSection extends StatelessWidget {
  final List<InspectionImage> existingImages;
  final List<XFile> pendingFiles;
  final InspectionImageRepository imageRepo;
  final int apiaryId;
  final int hiveId;
  final int inspectionId;
  final ValueChanged<InspectionImage> onDeleteExisting;
  final ValueChanged<XFile> onAddPending;
  final ValueChanged<XFile> onRemovePending;

  const InspectionImagesSection({
    super.key,
    required this.existingImages,
    required this.pendingFiles,
    required this.imageRepo,
    required this.apiaryId,
    required this.hiveId,
    required this.inspectionId,
    required this.onDeleteExisting,
    required this.onAddPending,
    required this.onRemovePending,
  });

  int get _total => existingImages.length + pendingFiles.length;
  bool get _atLimit => _total >= _kMaxPhotos;

  Future<void> _pickImage(BuildContext context) async {
    if (_atLimit) return;
    final picker = ImagePicker();
    final ImageSource source;
    if (kIsWeb) {
      source = ImageSource.gallery;
    } else {
      final chosen = await _chooseSource(context);
      if (chosen == null) return;
      source = chosen;
    }
    final file = await picker.pickImage(source: source, imageQuality: 85);
    if (file == null) return;
    if (!context.mounted) return;
    final sizeError = await validateImageFileSize(
      file,
      AppLocalizations.of(context)!,
    );
    if (sizeError != null) {
      if (context.mounted) showBigSnackBar(context, sizeError);
      return;
    }
    onAddPending(file);
  }

  Future<ImageSource?> _chooseSource(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return showModalBottomSheet<ImageSource>(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.photo_library_outlined),
              title: Text(l10n.inspectionPhotoSourceGallery),
              onTap: () => Navigator.of(context).pop(ImageSource.gallery),
            ),
            ListTile(
              leading: const Icon(Icons.camera_alt_outlined),
              title: Text(l10n.inspectionPhotoSourceCamera),
              onTap: () => Navigator.of(context).pop(ImageSource.camera),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _confirmDeleteExisting(
    BuildContext context,
    InspectionImage img,
  ) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.inspectionDeletePhoto),
        content: Text(l10n.inspectionDeletePhotoWarning),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: Text(l10n.generalCancel),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            child: Text(
              l10n.generalDelete,
              style: TextStyle(color: Theme.of(ctx).colorScheme.error),
            ),
          ),
        ],
      ),
    );
    if (confirmed ?? false) onDeleteExisting(img);
  }

  void _openViewer(BuildContext context, int index) {
    final headers = imageRepo.authHeaders();
    Navigator.of(context).push(
      PageRouteBuilder(
        opaque: false,
        barrierColor: Colors.black87,
        pageBuilder: (context, anim, secondaryAnim) => InspectionPhotoViewer(
          existingImages: existingImages,
          pendingFiles: pendingFiles,
          initialIndex: index,
          urlBuilder: (img) =>
              imageRepo.imageUrl(apiaryId, hiveId, inspectionId, img.id),
          headers: headers,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final headers = imageRepo.authHeaders();
    final hasAny = _total > 0;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              '${l10n.inspectionPhotos}  $_total/$_kMaxPhotos',
              style: Theme.of(context).textTheme.titleSmall?.copyWith(
                    color: Theme.of(context).colorScheme.primary,
                  ),
            ),
            TextButton.icon(
              onPressed: _atLimit ? null : () => _pickImage(context),
              icon: const Icon(Icons.add_photo_alternate_outlined, size: 18),
              label: Text(l10n.inspectionAddPhoto),
            ),
          ],
        ),
        if (hasAny)
          Center(
            child: SizedBox(
              height: 100,
              child: ListView(
                shrinkWrap: true,
                scrollDirection: Axis.horizontal,
                children: [
                  for (int i = 0; i < existingImages.length; i++)
                    _ExistingThumbnail(
                      url: imageRepo.imageUrl(
                          apiaryId, hiveId, inspectionId, existingImages[i].id),
                      headers: headers,
                      onTap: () => _openViewer(context, i),
                      onDelete: () =>
                          _confirmDeleteExisting(context, existingImages[i]),
                    ),
                  for (int i = 0; i < pendingFiles.length; i++)
                    _PendingThumbnail(
                      file: pendingFiles[i],
                      onTap: () =>
                          _openViewer(context, existingImages.length + i),
                      onRemove: () => onRemovePending(pendingFiles[i]),
                    ),
                ],
              ),
            ),
          ),
      ],
    );
  }
}

class _ExistingThumbnail extends StatelessWidget {
  final String url;
  final Map<String, String> headers;
  final VoidCallback onTap;
  final VoidCallback onDelete;

  const _ExistingThumbnail({
    required this.url,
    required this.headers,
    required this.onTap,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    return _ThumbnailFrame(
      onTap: onTap,
      onDelete: onDelete,
      child: Image.network(
        url,
        headers: headers,
        fit: BoxFit.cover,
        errorBuilder: (context, err, stack) =>
            const Icon(Icons.broken_image_outlined),
      ),
    );
  }
}

class _PendingThumbnail extends StatelessWidget {
  final XFile file;
  final VoidCallback onTap;
  final VoidCallback onRemove;

  const _PendingThumbnail({
    required this.file,
    required this.onTap,
    required this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    return _ThumbnailFrame(
      onTap: onTap,
      onDelete: onRemove,
      pending: true,
      child: FutureBuilder<Uint8List>(
        future: file.readAsBytes(),
        builder: (_, snapshot) {
          if (snapshot.hasData) {
            return Image.memory(snapshot.data!, fit: BoxFit.cover);
          }
          return const Center(child: CircularProgressIndicator(strokeWidth: 2));
        },
      ),
    );
  }
}

class _ThumbnailFrame extends StatelessWidget {
  final Widget child;
  final VoidCallback onTap;
  final VoidCallback onDelete;
  final bool pending;

  const _ThumbnailFrame({
    required this.child,
    required this.onTap,
    required this.onDelete,
    this.pending = false,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: 8, bottom: 4),
      child: GestureDetector(
        onTap: onTap,
        child: Stack(
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: SizedBox(width: 90, height: 90, child: child),
            ),
            if (pending)
              Positioned(
                top: 2,
                left: 2,
                child: Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                  decoration: BoxDecoration(
                    color: Theme.of(context).colorScheme.primaryContainer,
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    'NEW',
                    style: Theme.of(context).textTheme.labelSmall?.copyWith(
                          color: Theme.of(context)
                              .colorScheme
                              .onPrimaryContainer,
                          fontSize: 8,
                        ),
                  ),
                ),
              ),
            Positioned(
              top: 2,
              right: 2,
              child: GestureDetector(
                onTap: onDelete,
                child: Container(
                  decoration: BoxDecoration(
                    color: Colors.black54,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child:
                      const Icon(Icons.close, color: Colors.white, size: 18),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ── Photo viewer ─────────────────────────────────────────────────────────────

class InspectionPhotoViewer extends StatefulWidget {
  final List<InspectionImage> existingImages;
  final List<XFile> pendingFiles;
  final int initialIndex;
  final String Function(InspectionImage) urlBuilder;
  final Map<String, String> headers;

  const InspectionPhotoViewer({
    required this.existingImages,
    required this.pendingFiles,
    required this.initialIndex,
    required this.urlBuilder,
    required this.headers,
  });

  @override
  State<InspectionPhotoViewer> createState() => _InspectionPhotoViewerState();
}

class _InspectionPhotoViewerState extends State<InspectionPhotoViewer> {
  late final PageController _controller;
  late int _current;

  int get _total =>
      widget.existingImages.length + widget.pendingFiles.length;

  @override
  void initState() {
    super.initState();
    _current = widget.initialIndex;
    _controller = PageController(initialPage: widget.initialIndex);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      appBar: AppBar(
        backgroundColor: Colors.black54,
        foregroundColor: Colors.white,
        title: Text('${_current + 1} / $_total',
            style: const TextStyle(color: Colors.white)),
        leading: IconButton(
          icon: const Icon(Icons.close, color: Colors.white),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: PageView.builder(
        controller: _controller,
        itemCount: _total,
        onPageChanged: (i) => setState(() => _current = i),
        itemBuilder: (_, i) {
          if (i < widget.existingImages.length) {
            final img = widget.existingImages[i];
            return InteractiveViewer(
              child: Center(
                child: Image.network(
                  widget.urlBuilder(img),
                  headers: widget.headers,
                  fit: BoxFit.contain,
                  errorBuilder: (context, err, stack) => const Icon(
                      Icons.broken_image_outlined,
                      color: Colors.white,
                      size: 64),
                ),
              ),
            );
          }
          final file = widget.pendingFiles[i - widget.existingImages.length];
          return FutureBuilder<Uint8List>(
            future: file.readAsBytes(),
            builder: (_, snap) {
              if (snap.hasData) {
                return InteractiveViewer(
                  child: Center(
                    child: Image.memory(snap.data!, fit: BoxFit.contain),
                  ),
                );
              }
              return const Center(
                  child: CircularProgressIndicator(color: Colors.white));
            },
          );
        },
      ),
    );
  }
}
