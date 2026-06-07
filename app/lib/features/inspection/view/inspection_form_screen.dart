import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import '../../../core/widgets/profile_icon_button.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:image_picker/image_picker.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../l10n/app_localizations.dart';
import '../../hive/data/hive_model.dart';
import '../../hive/data/hive_repository.dart';
import '../../hive/view/hive_form_widgets.dart';
import '../data/inspection_image_model.dart';
import '../data/inspection_image_repository.dart';
import '../data/inspection_model.dart';
import '../data/inspection_repository.dart';

class InspectionFormScreen extends StatefulWidget {
  final int apiaryId;
  final Hive hive;
  final Inspection? inspection;
  final Inspection? previousInspection;

  const InspectionFormScreen({
    super.key,
    required this.apiaryId,
    required this.hive,
    this.inspection,
    this.previousInspection,
  });

  bool get isEditing => inspection != null;

  @override
  State<InspectionFormScreen> createState() => _InspectionFormScreenState();
}

class _InspectionFormScreenState extends State<InspectionFormScreen> {
  final _formKey = GlobalKey<FormState>();

  late DateTime _inspectedAt;
  late bool _queenSeen;
  late String _broodPattern;
  late String _aggressiveness;
  late bool _queenAdded;
  late final TextEditingController _framesBroodController;
  late final TextEditingController _framesHoneyController;
  late final TextEditingController _framesPollenController;
  late final TextEditingController _framesAddedDrawnController;
  late final TextEditingController _framesAddedFoundationController;
  late final TextEditingController _framesAddedHoneyController;
  late final TextEditingController _queenCellsCountController;
  late final TextEditingController _notesController;

  late bool _hiveActive;
  late bool _hiveQueenless;
  late bool _hiveReadyForHarvest;
  late Set<String> _hiveDiseases;

  List<InspectionImage> _existingImages = [];
  final List<XFile> _pendingImages = [];

  bool _loading = false;

  @override
  void initState() {
    super.initState();
    final insp = widget.inspection;
    final prev = !widget.isEditing ? widget.previousInspection : null;

    _inspectedAt = insp?.inspectedAt ?? DateTime.now();
    _queenSeen = insp != null ? insp.queenSeen == 'seen' : false;
    _broodPattern = insp?.broodPattern ?? '';
    _aggressiveness = insp?.aggressiveness ?? '';
    _queenAdded = insp?.queenAdded ?? false;

    _framesBroodController = _initFrameCtrl(insp?.framesBrood, prev?.framesBrood);
    _framesHoneyController = _initFrameCtrl(insp?.framesHoney, prev?.framesHoney);
    _framesPollenController = _initFrameCtrl(insp?.framesPollen, prev?.framesPollen);

    _framesAddedDrawnController = TextEditingController(
      text: insp?.framesAddedDrawn?.toString() ?? '0',
    );
    _framesAddedFoundationController = TextEditingController(
      text: insp?.framesAddedFoundation?.toString() ?? '0',
    );
    _framesAddedHoneyController = TextEditingController(
      text: insp?.framesAddedHoney?.toString() ?? '0',
    );

    void rebuild() => setState(() {});
    _framesBroodController.addListener(rebuild);
    _framesHoneyController.addListener(rebuild);
    _framesPollenController.addListener(rebuild);
    _framesAddedDrawnController.addListener(rebuild);
    _framesAddedFoundationController.addListener(rebuild);
    _framesAddedHoneyController.addListener(rebuild);
    _queenCellsCountController = TextEditingController(
      text: insp?.queenCellsCount?.toString() ?? '',
    );
    _notesController = TextEditingController(text: insp?.notes ?? '');

    _hiveActive = widget.hive.active;
    if (widget.isEditing) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _loadImages());
    }
    _hiveQueenless = widget.hive.queenless;
    _hiveReadyForHarvest = widget.hive.readyForHarvest;
    _hiveDiseases = widget.hive.diseases.map((d) => d.disease).toSet();
  }

  TextEditingController _initFrameCtrl(int? editValue, int? prevValue) {
    final text = editValue?.toString() ?? prevValue?.toString() ?? '';
    return TextEditingController(text: text);
  }

  @override
  void dispose() {
    _framesBroodController.dispose();
    _framesHoneyController.dispose();
    _framesPollenController.dispose();
    _framesAddedDrawnController.dispose();
    _framesAddedFoundationController.dispose();
    _framesAddedHoneyController.dispose();
    _queenCellsCountController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  Future<void> _deleteExistingImage(InspectionImage img) async {
    final api = context.read<ApiClient>();
    try {
      await InspectionImageRepository(api: api).deleteImage(
        widget.apiaryId,
        widget.hive.id,
        widget.inspection!.id,
        img.id,
      );
      if (mounted) setState(() => _existingImages.remove(img));
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.generalError)),
        );
      }
    }
  }

  Future<void> _loadImages() async {
    final api = context.read<ApiClient>();
    final repo = InspectionImageRepository(api: api);
    try {
      final imgs = await repo.listImages(
        widget.apiaryId,
        widget.hive.id,
        widget.inspection!.id,
      );
      if (mounted) setState(() => _existingImages = imgs);
    } catch (_) {}
  }

  int? _parseOptionalInt(String text) {
    final trimmed = text.trim();
    if (trimmed.isEmpty) return null;
    return int.tryParse(trimmed);
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _inspectedAt,
      firstDate: DateTime(2000),
      lastDate: DateTime.now(),
    );
    if (picked != null && mounted) {
      setState(
        () => _inspectedAt = DateTime(
          picked.year,
          picked.month,
          picked.day,
          _inspectedAt.hour,
          _inspectedAt.minute,
        ),
      );
    }
  }

  Future<void> _submit(BuildContext ctx) async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() => _loading = true);
    final api = ctx.read<ApiClient>();
    final inspRepo = InspectionRepository(api: api);
    final imageRepo = InspectionImageRepository(api: api);
    final hiveRepo = HiveRepository(api: api);
    try {
      int inspectionId;
      if (widget.isEditing) {
        inspectionId = widget.inspection!.id;
        await inspRepo.updateInspection(
          apiaryId: widget.apiaryId,
          hiveId: widget.hive.id,
          inspectionId: inspectionId,
          inspectedAt: _inspectedAt,
          queenSeen: _queenSeen ? 'seen' : 'not_seen',
          broodPattern: _broodPattern,
          aggressiveness: _aggressiveness,
          queenAdded: _queenAdded,
          notes: _notesController.text.trim(),
          framesBrood: _parseOptionalInt(_framesBroodController.text),
          framesHoney: _parseOptionalInt(_framesHoneyController.text),
          framesPollen: _parseOptionalInt(_framesPollenController.text),
          framesAddedDrawn: _parseOptionalInt(_framesAddedDrawnController.text),
          framesAddedFoundation: _parseOptionalInt(
            _framesAddedFoundationController.text,
          ),
          framesAddedHoney: _parseOptionalInt(_framesAddedHoneyController.text),
          queenCellsCount: _parseOptionalInt(_queenCellsCountController.text),
        );
      } else {
        final created = await inspRepo.createInspection(
          apiaryId: widget.apiaryId,
          hiveId: widget.hive.id,
          inspectedAt: _inspectedAt,
          queenSeen: _queenSeen ? 'seen' : 'not_seen',
          broodPattern: _broodPattern,
          aggressiveness: _aggressiveness,
          queenAdded: _queenAdded,
          notes: _notesController.text.trim(),
          framesBrood: _parseOptionalInt(_framesBroodController.text),
          framesHoney: _parseOptionalInt(_framesHoneyController.text),
          framesPollen: _parseOptionalInt(_framesPollenController.text),
          framesAddedDrawn: _parseOptionalInt(_framesAddedDrawnController.text),
          framesAddedFoundation: _parseOptionalInt(
            _framesAddedFoundationController.text,
          ),
          framesAddedHoney: _parseOptionalInt(_framesAddedHoneyController.text),
          queenCellsCount: _parseOptionalInt(_queenCellsCountController.text),
        );
        inspectionId = created.id;
      }
      for (final file in _pendingImages) {
        await imageRepo.uploadImage(
          widget.apiaryId,
          widget.hive.id,
          inspectionId,
          file,
        );
      }
      if (!ctx.mounted) return;
      await _syncHiveState(ctx, hiveRepo);
      if (ctx.mounted) Navigator.of(ctx).pop(true);
    } catch (_) {
      if (ctx.mounted) {
        ScaffoldMessenger.of(ctx).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(ctx)!.generalError)),
        );
      }
      setState(() => _loading = false);
    }
  }

  int get _totalFramesAdded =>
      (int.tryParse(_framesAddedDrawnController.text) ?? 0) +
      (int.tryParse(_framesAddedFoundationController.text) ?? 0) +
      (int.tryParse(_framesAddedHoneyController.text) ?? 0);

  bool get _framesWarning {
    if (widget.hive.frames <= 0) return false;
    final capacity = widget.hive.frames + _totalFramesAdded;
    final inHive = (int.tryParse(_framesBroodController.text) ?? 0) +
        (int.tryParse(_framesHoneyController.text) ?? 0) +
        (int.tryParse(_framesPollenController.text) ?? 0);
    return inHive > capacity;
  }

  Future<void> _syncHiveState(BuildContext ctx, HiveRepository hiveRepo) async {
    if (_hiveActive != widget.hive.active ||
        _hiveQueenless != widget.hive.queenless ||
        _hiveReadyForHarvest != widget.hive.readyForHarvest) {
      await hiveRepo.updateHive(
        apiaryId: widget.apiaryId,
        hiveId: widget.hive.id,
        name: widget.hive.name,
        type: widget.hive.type,
        active: _hiveActive,
        queenless: _hiveQueenless,
        readyForHarvest: _hiveReadyForHarvest,
        frames: widget.hive.frames,
      );
    }

    if (!widget.isEditing && _totalFramesAdded > 0) {
      await hiveRepo.addFrames(
        apiaryId: widget.apiaryId,
        hiveId: widget.hive.id,
        delta: _totalFramesAdded,
      );
    }

    final existing = widget.hive.diseases.map((d) => d.disease).toSet();
    final toAdd = _hiveDiseases.difference(existing);
    final toRemove = existing.difference(_hiveDiseases);

    for (final disease in toAdd) {
      await hiveRepo.addDisease(
        apiaryId: widget.apiaryId,
        hiveId: widget.hive.id,
        disease: disease,
      );
    }
    for (final disease in toRemove) {
      final d = widget.hive.diseases.firstWhere((d) => d.disease == disease);
      await hiveRepo.removeDisease(
        apiaryId: widget.apiaryId,
        hiveId: widget.hive.id,
        diseaseId: d.id,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final hasPhotos = _existingImages.isNotEmpty || _pendingImages.isNotEmpty;
    final imageRepo = InspectionImageRepository(api: context.read<ApiClient>());

    return Scaffold(
      appBar: AppBar(
        title: Text(
          widget.isEditing ? l10n.inspectionEdit : l10n.inspectionAdd,
        ),
        actions: const [ProfileIconButton()],
      ),
      body: Column(
        children: [
          Expanded(
            child: SafeArea(
              bottom: false,
              child: Center(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(24),
                  child: ConstrainedBox(
                    constraints: AppLayout.formConstraints(context),
                    child: Form(
                      key: _formKey,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          _DateField(
                            date: _inspectedAt,
                            label: l10n.inspectionDate,
                            onTap: _pickDate,
                          ),
                          const SizedBox(height: 20),
                          _SectionTitle(l10n.inspectionSectionObservations),
                          const SizedBox(height: 12),
                          _BoolRow(
                            label: l10n.inspectionQueenSeen,
                            value: _queenSeen,
                            onChanged: (v) => setState(() => _queenSeen = v),
                          ),
                          const SizedBox(height: 16),
                          _EnumDropdown(
                            label: l10n.inspectionBroodPattern,
                            value: _broodPattern.isEmpty ? null : _broodPattern,
                            items: broodPatternValues,
                            labelFor: (v) => _broodPatternLabel(l10n, v),
                            onChanged: (v) =>
                                setState(() => _broodPattern = v ?? ''),
                          ),
                          const SizedBox(height: 16),
                          _EnumDropdown(
                            label: l10n.inspectionAggressiveness,
                            value: _aggressiveness.isEmpty
                                ? null
                                : _aggressiveness,
                            items: aggressivenessValues,
                            labelFor: (v) => _aggressivenessLabel(l10n, v),
                            onChanged: (v) =>
                                setState(() => _aggressiveness = v ?? ''),
                          ),
                          const SizedBox(height: 16),
                          _NumericField(
                            controller: _queenCellsCountController,
                            label: l10n.inspectionQueenCellsCount,
                          ),
                          const SizedBox(height: 20),
                          _SectionTitle(l10n.inspectionSectionFrames),
                          const SizedBox(height: 12),
                          if (_framesWarning) ...[
                            Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 12, vertical: 8),
                              decoration: BoxDecoration(
                                color: Theme.of(context)
                                    .colorScheme
                                    .errorContainer,
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: Row(
                                children: [
                                  Icon(
                                    Icons.warning_amber_rounded,
                                    size: 18,
                                    color: Theme.of(context)
                                        .colorScheme
                                        .onErrorContainer,
                                  ),
                                  const SizedBox(width: 8),
                                  Expanded(
                                    child: Text(
                                      l10n.hiveFramesWarning,
                                      style: Theme.of(context)
                                          .textTheme
                                          .bodySmall
                                          ?.copyWith(
                                            color: Theme.of(context)
                                                .colorScheme
                                                .onErrorContainer,
                                          ),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 12),
                          ],
                          _NumericField(
                            controller: _framesBroodController,
                            label: l10n.inspectionFramesBrood,
                          ),
                          const SizedBox(height: 16),
                          _NumericField(
                            controller: _framesHoneyController,
                            label: l10n.inspectionFramesHoney,
                          ),
                          const SizedBox(height: 16),
                          _NumericField(
                            controller: _framesPollenController,
                            label: l10n.inspectionFramesPollen,
                          ),
                          const SizedBox(height: 16),
                          _NumericField(
                            controller: _framesAddedDrawnController,
                            label: l10n.inspectionFramesAddedDrawn,
                          ),
                          const SizedBox(height: 16),
                          _NumericField(
                            controller: _framesAddedFoundationController,
                            label: l10n.inspectionFramesAddedFoundation,
                          ),
                          const SizedBox(height: 16),
                          _NumericField(
                            controller: _framesAddedHoneyController,
                            label: l10n.inspectionFramesAddedHoney,
                          ),
                          const SizedBox(height: 16),
                          TextFormField(
                            controller: _notesController,
                            decoration: InputDecoration(
                              labelText: l10n.inspectionNotes,
                              border: const OutlineInputBorder(),
                              alignLabelWithHint: true,
                            ),
                            maxLines: 4,
                            textInputAction: TextInputAction.newline,
                          ),
                          const SizedBox(height: 20),
                          _SectionTitle(l10n.inspectionSectionHiveState),
                          const SizedBox(height: 12),
                          _BoolRow(
                            label: l10n.hiveActive,
                            value: _hiveActive,
                            onChanged: (v) => setState(() => _hiveActive = v),
                          ),
                          const SizedBox(height: 12),
                          _BoolRow(
                            label: l10n.hiveQueenless,
                            value: _hiveQueenless,
                            onChanged: (v) =>
                                setState(() => _hiveQueenless = v),
                          ),
                          const SizedBox(height: 12),
                          _BoolRow(
                            label: l10n.hiveReadyForHarvest,
                            value: _hiveReadyForHarvest,
                            onChanged: (v) =>
                                setState(() => _hiveReadyForHarvest = v),
                          ),
                          const SizedBox(height: 12),
                          _BoolRow(
                            label: l10n.inspectionQueenAdded,
                            value: _queenAdded,
                            onChanged: (v) =>
                                setState(() => _queenAdded = v),
                          ),
                          const SizedBox(height: 12),
                          HiveDiseasesSection(
                            label: l10n.inspectionDiseases,
                            selected: _hiveDiseases,
                            onToggle: (disease, selected) {
                              setState(() {
                                if (selected) {
                                  _hiveDiseases = {
                                    ..._hiveDiseases,
                                    disease,
                                  };
                                } else {
                                  _hiveDiseases = _hiveDiseases
                                      .where((d) => d != disease)
                                      .toSet();
                                }
                              });
                            },
                          ),
                          // Photo gallery — only visible when photos exist
                          if (hasPhotos) ...[
                            const SizedBox(height: 16),
                            _FormPhotoGallery(
                              existingImages: _existingImages,
                              pendingFiles: _pendingImages,
                              imageRepo: imageRepo,
                              apiaryId: widget.apiaryId,
                              hiveId: widget.hive.id,
                              inspectionId: widget.inspection?.id ?? 0,
                              onDeleteExisting: _deleteExistingImage,
                              onRemovePending: (f) =>
                                  setState(() => _pendingImages.remove(f)),
                            ),
                          ],
                          const SizedBox(height: 24),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
          _InspectionFormBanner(
            loading: _loading,
            pendingCount: _pendingImages.length,
            atPhotoLimit:
                (_existingImages.length + _pendingImages.length) >= 6,
            onSave: () => _submit(context),
            onAddPhoto: _pickImageForBanner,
          ),
        ],
      ),
    );
  }

  Future<void> _pickImageForBanner() async {
    final total = _existingImages.length + _pendingImages.length;
    if (total >= 6) return;
    final picker = ImagePicker();
    final ImageSource source;
    if (kIsWeb) {
      source = ImageSource.gallery;
    } else {
      final chosen = await showModalBottomSheet<ImageSource>(
        context: context,
        builder: (_) => SafeArea(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              ListTile(
                leading: const Icon(Icons.photo_library_outlined),
                title: Text(AppLocalizations.of(context)!.inspectionPhotoSourceGallery),
                onTap: () => Navigator.of(context).pop(ImageSource.gallery),
              ),
              ListTile(
                leading: const Icon(Icons.camera_alt_outlined),
                title: Text(AppLocalizations.of(context)!.inspectionPhotoSourceCamera),
                onTap: () => Navigator.of(context).pop(ImageSource.camera),
              ),
            ],
          ),
        ),
      );
      if (chosen == null) return;
      source = chosen;
    }
    final file = await picker.pickImage(source: source, imageQuality: 85);
    if (file != null && mounted) setState(() => _pendingImages.add(file));
  }
}

// ── Amber banner: save + add-photo buttons ────────────────────────────────────

class _InspectionFormBanner extends StatelessWidget {
  final bool loading;
  final int pendingCount;
  final bool atPhotoLimit;
  final VoidCallback onSave;
  final VoidCallback onAddPhoto;

  const _InspectionFormBanner({
    required this.loading,
    required this.pendingCount,
    required this.atPhotoLimit,
    required this.onSave,
    required this.onAddPhoto,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final bannerWidth = AppLayout.bannerWidth(context);

    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.only(bottom: 8),
        child: Center(
          child: SizedBox(
            width: bannerWidth,
            child: Container(
              decoration: const BoxDecoration(
                color: Colors.amber,
                borderRadius: BorderRadius.only(
                  topLeft: Radius.circular(20),
                  topRight: Radius.circular(20),
                ),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black26,
                    blurRadius: 8,
                    offset: Offset(0, -2),
                  ),
                ],
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  // Add photo
                  Badge(
                    isLabelVisible: pendingCount > 0,
                    label: Text('$pendingCount'),
                    child: IconButton(
                      icon: const Icon(Icons.add_photo_alternate_outlined),
                      iconSize: 28,
                      tooltip: l10n.inspectionAddPhoto,
                      onPressed: atPhotoLimit ? null : onAddPhoto,
                    ),
                  ),
                  // Save
                  loading
                      ? const Padding(
                          padding: EdgeInsets.all(12),
                          child: SizedBox(
                            width: 24,
                            height: 24,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                        )
                      : IconButton(
                          icon: const Icon(Icons.check),
                          iconSize: 28,
                          tooltip: l10n.generalSave,
                          onPressed: onSave,
                        ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

// ── Inline photo gallery shown below diseases ─────────────────────────────────

class _FormPhotoGallery extends StatelessWidget {
  final List<InspectionImage> existingImages;
  final List<XFile> pendingFiles;
  final InspectionImageRepository imageRepo;
  final int apiaryId;
  final int hiveId;
  final int inspectionId;
  final ValueChanged<InspectionImage> onDeleteExisting;
  final ValueChanged<XFile> onRemovePending;

  const _FormPhotoGallery({
    required this.existingImages,
    required this.pendingFiles,
    required this.imageRepo,
    required this.apiaryId,
    required this.hiveId,
    required this.inspectionId,
    required this.onDeleteExisting,
    required this.onRemovePending,
  });

  void _openViewer(BuildContext context, int index) {
    Navigator.of(context).push(
      PageRouteBuilder(
        opaque: false,
        barrierColor: Colors.black87,
        pageBuilder: (ctx, anim, secondaryAnim) => _PhotoViewerRoute(
          existingImages: existingImages,
          pendingFiles: pendingFiles,
          initialIndex: index,
          urlBuilder: (img) =>
              imageRepo.imageUrl(apiaryId, hiveId, inspectionId, img.id),
          headers: imageRepo.authHeaders(),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final headers = imageRepo.authHeaders();
    final total = existingImages.length + pendingFiles.length;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          '${l10n.inspectionPhotos}  $total / 6',
          style: Theme.of(context).textTheme.titleSmall?.copyWith(
                color: Theme.of(context).colorScheme.primary,
              ),
        ),
        const SizedBox(height: 8),
        LayoutBuilder(
          builder: (context, constraints) {
            final size = kIsWeb ? 180.0 : 120.0;
            return SizedBox(
              height: size + 15,
              child: SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                child: ConstrainedBox(
                  constraints: BoxConstraints(minWidth: constraints.maxWidth),
                  child: Center(
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        for (int i = 0; i < existingImages.length; i++)
                          _GalleryThumb(
                            size: size,
                            child: Image.network(
                              imageRepo.imageUrl(apiaryId, hiveId, inspectionId,
                                  existingImages[i].id),
                              headers: headers,
                              fit: BoxFit.cover,
                              errorBuilder: (context, err, stack) =>
                                  const Icon(Icons.broken_image_outlined),
                            ),
                            onTap: () => _openViewer(context, i),
                            onDelete: () => onDeleteExisting(existingImages[i]),
                          ),
                        for (int i = 0; i < pendingFiles.length; i++)
                          _GalleryThumb(
                            size: size,
                            pending: true,
                            child: FutureBuilder<Uint8List>(
                              future: pendingFiles[i].readAsBytes(),
                              builder: (_, snap) => snap.hasData
                                  ? Image.memory(snap.data!, fit: BoxFit.cover)
                                  : const Center(
                                      child: CircularProgressIndicator(
                                          strokeWidth: 2)),
                            ),
                            onTap: () =>
                                _openViewer(context, existingImages.length + i),
                            onDelete: () => onRemovePending(pendingFiles[i]),
                          ),
                      ],
                    ),
                  ),
                ),
              ),
            );
          },
        ),
      ],
    );
  }
}

class _GalleryThumb extends StatelessWidget {
  final Widget child;
  final VoidCallback onTap;
  final VoidCallback onDelete;
  final double size;
  final bool pending;

  const _GalleryThumb({
    required this.child,
    required this.onTap,
    required this.onDelete,
    required this.size,
    this.pending = false,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: GestureDetector(
        onTap: onTap,
        child: Stack(
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: SizedBox(width: size, height: size, child: child),
            ),
            if (pending)
              Positioned(
                top: 2,
                left: 2,
                child: Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                  decoration: BoxDecoration(
                    color: Theme.of(context).colorScheme.primaryContainer,
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    'NEW',
                    style: TextStyle(
                      fontSize: 8,
                      fontWeight: FontWeight.w700,
                      color: Theme.of(context).colorScheme.onPrimaryContainer,
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
                  child: const Icon(Icons.close, color: Colors.white, size: 16),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ── Full-screen photo viewer ──────────────────────────────────────────────────

class _PhotoViewerRoute extends StatefulWidget {
  final List<InspectionImage> existingImages;
  final List<XFile> pendingFiles;
  final int initialIndex;
  final String Function(InspectionImage) urlBuilder;
  final Map<String, String> headers;

  const _PhotoViewerRoute({
    required this.existingImages,
    required this.pendingFiles,
    required this.initialIndex,
    required this.urlBuilder,
    required this.headers,
  });

  @override
  State<_PhotoViewerRoute> createState() => _PhotoViewerRouteState();
}

class _PhotoViewerRouteState extends State<_PhotoViewerRoute> {
  late final PageController _ctrl;
  late int _current;

  int get _total => widget.existingImages.length + widget.pendingFiles.length;

  @override
  void initState() {
    super.initState();
    _current = widget.initialIndex;
    _ctrl = PageController(initialPage: widget.initialIndex);
  }

  @override
  void dispose() {
    _ctrl.dispose();
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
      body: GestureDetector(
        onTap: () => Navigator.of(context).pop(),
        child: PageView.builder(
          controller: _ctrl,
          itemCount: _total,
          onPageChanged: (i) => setState(() => _current = i),
          itemBuilder: (_, i) {
            if (i < widget.existingImages.length) {
              return InteractiveViewer(
                child: Center(
                  child: Image.network(
                    widget.urlBuilder(widget.existingImages[i]),
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
              builder: (_, snap) => snap.hasData
                  ? InteractiveViewer(
                      child: Center(
                          child: Image.memory(snap.data!, fit: BoxFit.contain)))
                  : const Center(
                      child: CircularProgressIndicator(color: Colors.white)),
            );
          },
        ),
      ),
    );
  }
}

String _broodPatternLabel(AppLocalizations l10n, String v) => switch (v) {
  'none' => l10n.inspectionBroodNone,
  'poor' => l10n.inspectionBroodPoor,
  'good' => l10n.inspectionBroodGood,
  'excellent' => l10n.inspectionBroodExcellent,
  _ => v,
};

String _aggressivenessLabel(AppLocalizations l10n, String v) => switch (v) {
  'calm' => l10n.inspectionAggressivenessCalm,
  'mild' => l10n.inspectionAggressivenessMild,
  'aggressive' => l10n.inspectionAggressivenessAggressive,
  'very_aggressive' => l10n.inspectionAggressivenessVeryAggressive,
  _ => v,
};


class _SectionTitle extends StatelessWidget {
  final String text;
  const _SectionTitle(this.text);

  @override
  Widget build(BuildContext context) {
    return Text(
      text,
      style: Theme.of(context).textTheme.titleSmall?.copyWith(
        color: Theme.of(context).colorScheme.primary,
      ),
    );
  }
}

class _BoolRow extends StatelessWidget {
  final String label;
  final bool value;
  final ValueChanged<bool> onChanged;

  const _BoolRow({
    required this.label,
    required this.value,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label, style: Theme.of(context).textTheme.bodyMedium),
        Switch(value: value, onChanged: onChanged),
      ],
    );
  }
}

class _DateField extends StatelessWidget {
  final DateTime date;
  final String label;
  final VoidCallback onTap;

  const _DateField({
    required this.date,
    required this.label,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final formatted = DateFormat.yMd(
      Localizations.localeOf(context).toString(),
    ).format(date);
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(4),
      child: InputDecorator(
        decoration: InputDecoration(
          labelText: label,
          border: const OutlineInputBorder(),
          suffixIcon: const Icon(Icons.calendar_today_outlined, size: 18),
        ),
        child: Text(formatted),
      ),
    );
  }
}

class _EnumDropdown extends StatelessWidget {
  final String label;
  final String? value;
  final List<String> items;
  final String Function(String) labelFor;
  final void Function(String?) onChanged;

  const _EnumDropdown({
    required this.label,
    required this.value,
    required this.items,
    required this.labelFor,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return DropdownButtonFormField<String>(
      decoration: InputDecoration(
        labelText: label,
        border: const OutlineInputBorder(),
      ),
      value: value,
      hint: Text(
        l10n.inspectionNotSet,
        style: Theme.of(context).textTheme.bodyMedium?.copyWith(
          color: Theme.of(context).colorScheme.onSurfaceVariant,
        ),
      ),
      items: items
          .map((v) => DropdownMenuItem(value: v, child: Text(labelFor(v))))
          .toList(),
      onChanged: onChanged,
    );
  }
}

class _NumericField extends StatelessWidget {
  final TextEditingController controller;
  final String label;

  const _NumericField({
    required this.controller,
    required this.label,
  });

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: controller,
      decoration: InputDecoration(
        labelText: label,
        border: const OutlineInputBorder(),
      ),
      keyboardType: TextInputType.number,
      inputFormatters: [FilteringTextInputFormatter.digitsOnly],
    );
  }
}

