import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/storage/token_storage.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/delete_dialog.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../../inspection/data/inspection_image_repository.dart';
import '../../inspection/data/inspection_model.dart';
import '../../inspection/data/inspection_repository.dart';
import '../../inspection/view/inspection_form_screen.dart';
import '../../inspection/view/inspection_history_screen.dart';
import '../../inspection/view/inspection_images_section.dart';
import '../../inspection/view/inspection_summary.dart';
import '../../treatment/data/treatment_model.dart';
import '../../treatment/data/treatment_repository.dart';
import '../../harvest/data/harvest_model.dart';
import '../../harvest/data/harvest_repository.dart';
import '../../harvest/view/harvest_form_screen.dart';
import '../../harvest/view/harvest_history_screen.dart';
import '../../treatment/view/treatment_form_screen.dart';
import '../../treatment/view/treatment_history_screen.dart';
import '../../apiary/data/apiary_model.dart';
import '../../apiary/data/apiary_repository.dart';
import '../data/hive_model.dart';
import '../data/hive_repository.dart';
import 'change_apiary_modal.dart';
import 'edit_hive_screen.dart';
import 'hive_form_widgets.dart';

class HiveDetailScreen extends StatefulWidget {
  final Hive hive;
  final int apiaryId;

  const HiveDetailScreen({
    super.key,
    required this.hive,
    required this.apiaryId,
  });

  @override
  State<HiveDetailScreen> createState() => _HiveDetailScreenState();
}

class _HiveDetailScreenState extends State<HiveDetailScreen> {
  late Hive _hive;
  Inspection? _lastInspection;
  bool _inspectionLoaded = false;
  Treatment? _lastTreatment;
  bool _treatmentLoaded = false;
  Harvest? _lastHarvest;
  bool _harvestLoaded = false;
  List<Apiary>? _otherApiaries;

  @override
  void initState() {
    super.initState();
    _hive = widget.hive;
    _loadLastInspection();
    _loadLastTreatment();
    _loadLastHarvest();
    _loadOtherApiaries();
  }

  Future<void> _loadOtherApiaries() async {
    try {
      final all = await ApiaryRepository(api: context.read<ApiClient>())
          .listApiaries();
      if (mounted) {
        setState(() {
          _otherApiaries =
              all.where((a) => a.id != widget.apiaryId).toList();
        });
      }
    } catch (_) {}
  }

  Future<void> _refreshHive() async {
    try {
      final fresh = await HiveRepository(api: context.read<ApiClient>())
          .getHive(widget.apiaryId, _hive.id);
      if (mounted) setState(() => _hive = fresh);
    } catch (_) {}
  }

  Future<void> _loadLastInspection() async {
    try {
      final result = await InspectionRepository(
        api: context.read<ApiClient>(),
      ).listInspections(widget.apiaryId, _hive.id);
      if (!mounted) return;
      final inspections = result.items
        ..sort((a, b) => b.inspectedAt.compareTo(a.inspectedAt));
      setState(() {
        _inspectionLoaded = true;
        _lastInspection = inspections.isNotEmpty ? inspections.first : null;
      });
    } catch (_) {
      if (mounted) setState(() => _inspectionLoaded = true);
    }
  }

  Future<void> _openEdit() async {
    final updated = await Navigator.of(context).push<Hive>(
      MaterialPageRoute(
        builder: (_) => EditHiveScreen(apiaryId: widget.apiaryId, hive: _hive),
      ),
    );
    if (updated != null && mounted) {
      setState(() => _hive = updated);
    }
  }

  Future<void> _changeApiary() async {
    if (_otherApiaries == null || _otherApiaries!.isEmpty) return;
    final moved = await showDialog<bool>(
      context: context,
      builder: (_) => ChangeApiaryModal(
        apiaryId: widget.apiaryId,
        hiveId: _hive.id,
        apiClient: context.read<ApiClient>(),
        otherApiaries: _otherApiaries!,
      ),
    );
    if (moved == true && mounted) Navigator.of(context).pop();
  }

  Future<void> _delete() async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDeleteDialog(
      context,
      title: l10n.hiveDeleteConfirm,
      warning: l10n.hiveDeleteWarning,
      l10n: l10n,
      withPuzzle: _hive.lastInspectedAt != null,
    );
    if (!confirmed || !mounted) return;
    try {
      await HiveRepository(api: context.read<ApiClient>()).deleteHive(
        apiaryId: widget.apiaryId,
        hiveId: _hive.id,
      );
      if (mounted) Navigator.of(context).pop();
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
      }
    }
  }

  Future<void> _openInspections() async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => InspectionHistoryScreen(
          apiaryId: widget.apiaryId,
          hive: _hive,
        ),
      ),
    );
    if (!mounted) return;
    await _refreshHive();
    _loadLastInspection();
  }

  Future<void> _loadLastTreatment() async {
    try {
      final result = await TreatmentRepository(
        api: context.read<ApiClient>(),
      ).listTreatments(widget.apiaryId, _hive.id, limit: 1);
      if (!mounted) return;
      setState(() {
        _treatmentLoaded = true;
        _lastTreatment = result.items.isNotEmpty ? result.items.first : null;
      });
    } catch (_) {
      if (mounted) setState(() => _treatmentLoaded = true);
    }
  }

  Future<void> _openTreatments() async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => TreatmentHistoryScreen(
          apiaryId: widget.apiaryId,
          hive: _hive,
        ),
      ),
    );
    if (!mounted) return;
    _loadLastTreatment();
  }

  Future<void> _loadLastHarvest() async {
    try {
      final result = await HarvestRepository(
        api: context.read<ApiClient>(),
      ).listHarvests(widget.apiaryId, _hive.id, limit: 1);
      if (!mounted) return;
      setState(() {
        _harvestLoaded = true;
        _lastHarvest = result.items.isNotEmpty ? result.items.first : null;
      });
    } catch (_) {
      if (mounted) setState(() => _harvestLoaded = true);
    }
  }

  Future<void> _openHarvests() async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => HarvestHistoryScreen(
          apiaryId: widget.apiaryId,
          hive: _hive,
        ),
      ),
    );
    if (!mounted) return;
    _loadLastHarvest();
  }

  Future<void> _openCreateHarvest() async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => HarvestFormScreen(
          apiaryId: widget.apiaryId,
          hive: _hive,
        ),
      ),
    );
    if (!mounted) return;
    if (result == true) _loadLastHarvest();
  }

  Future<void> _openCreateTreatment() async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => TreatmentFormScreen(
          apiaryId: widget.apiaryId,
          hive: _hive,
        ),
      ),
    );
    if (!mounted) return;
    if (result == true) _loadLastTreatment();
  }

  Future<void> _openCreateInspection() async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => InspectionFormScreen(
          apiaryId: widget.apiaryId,
          hive: _hive,
          previousInspection: _lastInspection,
        ),
      ),
    );
    if (!mounted) return;
    if (result == true) await _refreshHive();
    _loadLastInspection();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(_hive.name),
        actions: const [ProfileIconButton()],
      ),
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: ConstrainedBox(
            constraints: AppLayout.formConstraints(context),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                _InfoCard(
                  hive: _hive,
                  onEdit: _openEdit,
                  onDelete: _delete,
                  onChangeApiary: (_otherApiaries != null &&
                          _otherApiaries!.isNotEmpty)
                      ? _changeApiary
                      : null,
                ),
                const SizedBox(height: 16),
                _InspectionSectionCard(
                  lastInspection: _lastInspection,
                  inspectionLoaded: _inspectionLoaded,
                  onAdd: _openCreateInspection,
                  onViewAll: _openInspections,
                  apiaryId: widget.apiaryId,
                ),
                const SizedBox(height: 16),
                _TreatmentSectionCard(
                  lastTreatment: _lastTreatment,
                  treatmentLoaded: _treatmentLoaded,
                  onAdd: _openCreateTreatment,
                  onViewAll: _openTreatments,
                ),
                const SizedBox(height: 16),
                _HarvestSectionCard(
                  lastHarvest: _lastHarvest,
                  harvestLoaded: _harvestLoaded,
                  onAdd: _openCreateHarvest,
                  onViewAll: _openHarvests,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

enum _HiveInfoAction { edit, changeApiary, delete }

class _InfoCard extends StatelessWidget {
  final Hive hive;
  final VoidCallback onEdit;
  final VoidCallback onDelete;
  final VoidCallback? onChangeApiary;

  const _InfoCard({
    required this.hive,
    required this.onEdit,
    required this.onDelete,
    this.onChangeApiary,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;

    final statusChips = <Widget>[
      _StatusChip(
        label: hive.active ? l10n.hiveActive : l10n.hiveInactive,
        background: hive.active
            ? Colors.green.shade100
            : colorScheme.errorContainer,
        foreground: hive.active
            ? Colors.green.shade800
            : colorScheme.onErrorContainer,
      ),
      if (hive.queenless)
        _StatusChip(
          label: l10n.hiveQueenless,
          background: colorScheme.errorContainer,
          foreground: colorScheme.onErrorContainer,
        ),
      if (hive.readyForHarvest)
        _StatusChip(
          label: l10n.hiveReadyForHarvest,
          background: Colors.amber.shade100,
          foreground: Colors.amber.shade900,
        ),
    ];

    return Card(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: Wrap(
                    spacing: 8,
                    runSpacing: 6,
                    children: [
                      _StatusChip(
                        label: hiveTypeLabels[hive.type] ?? hive.type,
                        background: colorScheme.secondaryContainer,
                        foreground: colorScheme.onSecondaryContainer,
                      ),
                    ],
                  ),
                ),
                PopupMenuButton<_HiveInfoAction>(
                  icon: const Icon(Icons.more_vert),
                  onSelected: (action) {
                    if (action == _HiveInfoAction.edit) onEdit();
                    if (action == _HiveInfoAction.changeApiary) {
                      onChangeApiary?.call();
                    }
                    if (action == _HiveInfoAction.delete) onDelete();
                  },
                  itemBuilder: (_) => [
                    PopupMenuItem(
                      value: _HiveInfoAction.edit,
                      child: ListTile(
                        leading: const Icon(Icons.edit_outlined),
                        title: Text(AppLocalizations.of(context)!.generalEdit),
                        contentPadding: EdgeInsets.zero,
                        visualDensity: VisualDensity.compact,
                      ),
                    ),
                    if (onChangeApiary != null)
                      PopupMenuItem(
                        value: _HiveInfoAction.changeApiary,
                        child: ListTile(
                          leading: const Icon(Icons.swap_horiz_outlined),
                          title: Text(
                            AppLocalizations.of(context)!.hiveChangeApiary,
                          ),
                          contentPadding: EdgeInsets.zero,
                          visualDensity: VisualDensity.compact,
                        ),
                      ),
                    PopupMenuItem(
                      value: _HiveInfoAction.delete,
                      child: ListTile(
                        leading: Icon(
                          Icons.delete_outline,
                          color: colorScheme.error,
                        ),
                        title: Text(
                          AppLocalizations.of(context)!.generalDelete,
                          style: TextStyle(color: colorScheme.error),
                        ),
                        contentPadding: EdgeInsets.zero,
                        visualDensity: VisualDensity.compact,
                      ),
                    ),
                  ],
                ),
              ],
            ),
            const Divider(height: 24),
            Text(
              l10n.hiveStatus,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: colorScheme.onSurfaceVariant,
                  ),
            ),
            const SizedBox(height: 6),
            Wrap(
              spacing: 8,
              runSpacing: 6,
              children: statusChips,
            ),
            if (hive.diseases.isNotEmpty) ...[
              const Divider(height: 24),
              Text(
                l10n.hiveDiseases,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: colorScheme.onSurfaceVariant,
                    ),
              ),
              const SizedBox(height: 6),
              Wrap(
                spacing: 8,
                runSpacing: 6,
                children: hive.diseases
                    .map((d) => _StatusChip(
                          label: hiveDiseaseLabel(l10n, d.disease),
                          background: colorScheme.errorContainer,
                          foreground: colorScheme.onErrorContainer,
                        ))
                    .toList(),
              ),
            ],
          ],
        ),
      ),
    );
  }
}


class _StatusChip extends StatelessWidget {
  final String label;
  final Color background;
  final Color foreground;

  const _StatusChip({
    required this.label,
    required this.background,
    required this.foreground,
  });

  @override
  Widget build(BuildContext context) {
    return Chip(
      label: Text(label),
      side: BorderSide.none,
      backgroundColor: background,
      labelStyle: TextStyle(color: foreground, fontSize: 12),
      visualDensity: VisualDensity.compact,
      padding: EdgeInsets.zero,
    );
  }
}

class _InspectionSectionCard extends StatelessWidget {
  final Inspection? lastInspection;
  final bool inspectionLoaded;
  final VoidCallback onAdd;
  final VoidCallback onViewAll;
  final int apiaryId;

  const _InspectionSectionCard({
    required this.lastInspection,
    required this.inspectionLoaded,
    required this.onAdd,
    required this.onViewAll,
    required this.apiaryId,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Card(
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InkWell(
            onTap: lastInspection != null ? onViewAll : null,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.search, size: 20, color: colorScheme.primary),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          l10n.hiveDetailInspections,
                          style: textTheme.titleMedium,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  if (!inspectionLoaded)
                    const SizedBox.shrink()
                  else if (lastInspection == null)
                    Text(
                      l10n.hiveDetailNoInspections,
                      style: textTheme.bodyMedium?.copyWith(
                        color: colorScheme.onSurfaceVariant,
                      ),
                    )
                  else
                    InspectionSummary(
                      inspection: lastInspection!,
                      showDate: true,
                      currentUserName: context.read<TokenStorage>().name,
                    ),
                ],
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
            child: Center(
              child: Wrap(
                spacing: 8,
                runSpacing: 8,
                alignment: WrapAlignment.center,
                children: [
                  if (lastInspection != null && lastInspection!.photoCount > 0)
                    OutlinedButton(
                      onPressed: () => showInspectionPhotosSheet(
                        context,
                        apiaryId: apiaryId,
                        hiveId: lastInspection!.hiveId,
                        inspection: lastInspection!,
                        imageRepo: InspectionImageRepository(
                          api: context.read<ApiClient>(),
                        ),
                      ),
                      child: Text(l10n.inspectionPhotos),
                    ),
                  OutlinedButton(
                    onPressed: onAdd,
                    child: Text(l10n.hiveDetailAddInspection),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _TreatmentSummary extends StatelessWidget {
  final Treatment treatment;
  final AppLocalizations l10n;
  final String? currentUserName;

  const _TreatmentSummary({
    required this.treatment,
    required this.l10n,
    this.currentUserName,
  });

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    final bodyStyle = textTheme.bodyMedium?.copyWith(
      color: colorScheme.onSurfaceVariant,
    );
    final labelStyle = textTheme.labelSmall?.copyWith(
      color: colorScheme.primary,
      fontWeight: FontWeight.w600,
      letterSpacing: 0.4,
    );

    final dateStr = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).add_Hm().format(treatment.treatedAt);

    final doseCount = int.tryParse(treatment.dose);
    final doseDisplay = doseCount != null
        ? l10n.treatmentDoseCount(doseCount)
        : treatment.dose;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          dateStr,
          style: textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
        ),
        if (treatment.treatedByName != null &&
            treatment.treatedByName != currentUserName) ...[
          const SizedBox(height: 2),
          Text(
            l10n.treatmentTreatedBy(treatment.treatedByName!),
            style: bodyStyle,
          ),
        ],
        const SizedBox(height: 8),
        Text(l10n.treatmentMedicine, style: labelStyle),
        const SizedBox(height: 2),
        Text(
          '${treatment.medicineName} · $doseDisplay',
          style: bodyStyle,
        ),
        if (treatment.notes.isNotEmpty) ...[
          const SizedBox(height: 8),
          Text(l10n.treatmentNote, style: labelStyle),
          const SizedBox(height: 2),
          Text(
            treatment.notes,
            style: bodyStyle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ],
    );
  }
}

class _TreatmentSectionCard extends StatelessWidget {
  final Treatment? lastTreatment;
  final bool treatmentLoaded;
  final VoidCallback onAdd;
  final VoidCallback onViewAll;

  const _TreatmentSectionCard({
    required this.lastTreatment,
    required this.treatmentLoaded,
    required this.onAdd,
    required this.onViewAll,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Card(
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InkWell(
            onTap: lastTreatment != null ? onViewAll : null,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(
                        Icons.medical_services_outlined,
                        size: 20,
                        color: colorScheme.primary,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          l10n.hiveDetailTreatments,
                          style: textTheme.titleMedium,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  if (!treatmentLoaded)
                    const SizedBox.shrink()
                  else if (lastTreatment == null)
                    Text(
                      l10n.hiveDetailNoTreatments,
                      style: textTheme.bodyMedium?.copyWith(
                        color: colorScheme.onSurfaceVariant,
                      ),
                    )
                  else
                    _TreatmentSummary(
                      treatment: lastTreatment!,
                      l10n: l10n,
                      currentUserName: context.read<TokenStorage>().name,
                    ),
                ],
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
            child: Center(
              child: Wrap(
                spacing: 8,
                runSpacing: 8,
                alignment: WrapAlignment.center,
                children: [
                  OutlinedButton(
                    onPressed: onAdd,
                    child: Text(l10n.hiveDetailLogTreatment),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _HarvestSummary extends StatelessWidget {
  final Harvest harvest;
  final AppLocalizations l10n;
  final String? currentUserName;

  const _HarvestSummary({
    required this.harvest,
    required this.l10n,
    this.currentUserName,
  });

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;
    final bodyStyle = textTheme.bodyMedium?.copyWith(
      color: colorScheme.onSurfaceVariant,
    );
    final labelStyle = textTheme.labelSmall?.copyWith(
      color: colorScheme.primary,
      fontWeight: FontWeight.w600,
      letterSpacing: 0.4,
    );
    final dateStr = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).add_Hm().format(harvest.harvestedAt);

    final otherHarvester =
        harvest.harvestedByName != null &&
                harvest.harvestedByName!.isNotEmpty &&
                harvest.harvestedByName != currentUserName
            ? harvest.harvestedByName
            : null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          dateStr,
          style: textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
        ),
        if (otherHarvester != null) ...[
          const SizedBox(height: 2),
          Text(
            l10n.harvestHarvestedBy(otherHarvester),
            style: bodyStyle,
          ),
        ],
        const SizedBox(height: 8),
        Text(l10n.harvestFrames, style: labelStyle),
        const SizedBox(height: 2),
        Text(
          '${harvest.frames} + ${harvest.halfFrames} ${l10n.harvestHalfFrames.toLowerCase()}',
          style: bodyStyle,
        ),
        const SizedBox(height: 8),
        Text(l10n.harvestKilograms, style: labelStyle),
        const SizedBox(height: 2),
        Text(
          '${harvest.kilograms.toStringAsFixed(2)} kg',
          style: bodyStyle,
        ),
        if (harvest.notes.isNotEmpty) ...[
          const SizedBox(height: 8),
          Text(l10n.harvestNote, style: labelStyle),
          const SizedBox(height: 2),
          Text(
            harvest.notes,
            style: bodyStyle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ],
    );
  }
}

class _HarvestSectionCard extends StatelessWidget {
  final Harvest? lastHarvest;
  final bool harvestLoaded;
  final VoidCallback onAdd;
  final VoidCallback onViewAll;

  const _HarvestSectionCard({
    required this.lastHarvest,
    required this.harvestLoaded,
    required this.onAdd,
    required this.onViewAll,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Card(
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InkWell(
            onTap: lastHarvest != null ? onViewAll : null,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(
                        Icons.water_drop_outlined,
                        size: 20,
                        color: colorScheme.primary,
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          l10n.hiveDetailHarvests,
                          style: textTheme.titleMedium,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  if (!harvestLoaded)
                    const SizedBox.shrink()
                  else if (lastHarvest == null)
                    Text(
                      l10n.hiveDetailNoHarvests,
                      style: textTheme.bodyMedium?.copyWith(
                        color: colorScheme.onSurfaceVariant,
                      ),
                    )
                  else
                    _HarvestSummary(
                      harvest: lastHarvest!,
                      l10n: l10n,
                      currentUserName: context.read<TokenStorage>().name,
                    ),
                ],
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
            child: Center(
              child: Wrap(
                spacing: 8,
                runSpacing: 8,
                alignment: WrapAlignment.center,
                children: [
                  OutlinedButton(
                    onPressed: onAdd,
                    child: Text(l10n.hiveDetailLogHarvest),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _SectionCard extends StatelessWidget {
  final String title;
  final IconData icon;
  final String emptyText;
  final String actionLabel;
  final VoidCallback? onAction;

  const _SectionCard({
    required this.title,
    required this.icon,
    required this.emptyText,
    required this.actionLabel,
    this.onAction,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, size: 20, color: colorScheme.primary),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(title, style: textTheme.titleMedium),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Text(
              emptyText,
              style: textTheme.bodyMedium?.copyWith(
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 12),
            Center(
              child: OutlinedButton(
                onPressed: onAction,
                child: Text(actionLabel),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
