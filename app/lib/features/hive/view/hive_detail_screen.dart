import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../../inspection/data/inspection_model.dart';
import '../../inspection/data/inspection_repository.dart';
import '../../inspection/view/inspection_form_screen.dart';
import '../../inspection/view/inspection_history_screen.dart';
import '../../inspection/view/inspection_summary.dart';
import '../data/hive_model.dart';
import '../data/hive_repository.dart';
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

  @override
  void initState() {
    super.initState();
    _hive = widget.hive;
    _loadLastInspection();
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
      final inspections = await InspectionRepository(
        api: context.read<ApiClient>(),
      ).listInspections(widget.apiaryId, _hive.id);
      if (!mounted) return;
      inspections.sort((a, b) => b.inspectedAt.compareTo(a.inspectedAt));
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
        actions: [
          const ProfileIconButton(),
          IconButton(
            icon: const Icon(Icons.edit_outlined),
            onPressed: _openEdit,
          ),
        ],
      ),
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: ConstrainedBox(
            constraints: AppLayout.formConstraints(context),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                _InfoCard(hive: _hive),
                const SizedBox(height: 16),
                _InspectionSectionCard(
                  lastInspection: _lastInspection,
                  inspectionLoaded: _inspectionLoaded,
                  onAdd: _openCreateInspection,
                  onViewAll: _openInspections,
                ),
                const SizedBox(height: 16),
                _SectionCard(
                  title: l10n.hiveDetailTreatments,
                  icon: Icons.medical_services_outlined,
                  emptyText: l10n.hiveDetailNoTreatments,
                  actionLabel: l10n.hiveDetailLogTreatment,
                  onAction: null,
                ),
                const SizedBox(height: 16),
                _SectionCard(
                  title: l10n.hiveDetailHarvests,
                  icon: Icons.water_drop_outlined,
                  emptyText: l10n.hiveDetailNoHarvests,
                  actionLabel: l10n.hiveDetailLogHarvest,
                  onAction: null,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _InfoCard extends StatelessWidget {
  final Hive hive;

  const _InfoCard({required this.hive});

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
            Wrap(
              spacing: 8,
              runSpacing: 6,
              children: [
                _StatusChip(
                  label: hiveTypeLabels[hive.type] ?? hive.type,
                  background: colorScheme.secondaryContainer,
                  foreground: colorScheme.onSecondaryContainer,
                ),
                if (hive.frames > 0)
                  _StatusChip(
                    label: '${l10n.hiveFrames}: ${hive.frames}',
                    background: colorScheme.secondaryContainer,
                    foreground: colorScheme.onSecondaryContainer,
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

  const _InspectionSectionCard({
    required this.lastInspection,
    required this.inspectionLoaded,
    required this.onAdd,
    required this.onViewAll,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
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
              InspectionSummary(inspection: lastInspection!, showDate: true),
            const SizedBox(height: 12),
            Center(
              child: Wrap(
                spacing: 8,
                runSpacing: 8,
                alignment: WrapAlignment.center,
                children: [
                  if (lastInspection != null)
                    OutlinedButton(
                      onPressed: onViewAll,
                      child: Text(l10n.hiveDetailViewInspections),
                    ),
                  OutlinedButton(
                    onPressed: onAdd,
                    child: Text(l10n.hiveDetailAddInspection),
                  ),
                ],
              ),
            ),
          ],
        ),
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
            OutlinedButton(
              onPressed: onAction,
              child: Text(actionLabel),
            ),
          ],
        ),
      ),
    );
  }
}
