import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';
import 'package:printing/printing.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../../feeding/data/feeding_model.dart';
import '../../feeding/data/feeding_repository.dart';
import '../../feeding/view/feeding_form_screen.dart';
import '../../feeding/view/feeding_history_screen.dart' show FeedingCard;
import '../../harvest/data/harvest_model.dart';
import '../../harvest/data/harvest_repository.dart';
import '../../harvest/view/harvest_form_screen.dart';
import '../../harvest/view/harvest_history_screen.dart' show HarvestCard;
import '../../hive/data/hive_model.dart';
import '../../inspection/data/inspection_model.dart';
import '../../inspection/data/inspection_repository.dart';
import '../../inspection/view/inspection_form_screen.dart';
import '../../inspection/view/inspection_history_screen.dart' show InspectionCard;
import '../../treatment/data/treatment_model.dart';
import '../../treatment/data/treatment_repository.dart';
import '../../treatment/view/treatment_form_screen.dart';
import '../../treatment/view/treatment_history_screen.dart' show TreatmentCard;

enum _ReportCategory { inspections, feedings, treatments, harvests }

class _ReportEntry {
  final _ReportCategory category;
  final DateTime date;
  final Object record;

  const _ReportEntry({
    required this.category,
    required this.date,
    required this.record,
  });
}

class DashboardScreen extends StatefulWidget {
  final int apiaryId;
  final List<Hive> hives;

  const DashboardScreen({
    super.key,
    required this.apiaryId,
    required this.hives,
  });

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> {
  late final List<Hive> _activeHives;
  late Set<int> _selectedHiveIds;
  Set<_ReportCategory> _selectedCategories = {
    _ReportCategory.inspections,
    _ReportCategory.feedings,
    _ReportCategory.treatments,
    _ReportCategory.harvests,
  };
  late DateTimeRange _range;

  bool _loading = false;
  bool _downloadingPdf = false;
  Map<int, List<_ReportEntry>>? _report;

  @override
  void initState() {
    super.initState();
    _activeHives = widget.hives.where((h) => h.active).toList();
    _selectedHiveIds = _activeHives.map((h) => h.id).toSet();
    final today = DateTime.now();
    final todayDate = DateTime(today.year, today.month, today.day);
    _range = DateTimeRange(
      start: todayDate.subtract(const Duration(days: 14)),
      end: todayDate,
    );
  }

  Hive _hiveById(int id) => _activeHives.firstWhere((h) => h.id == id);

  void _toggleHive(int hiveId, bool selected) {
    setState(() {
      if (selected) {
        _selectedHiveIds.add(hiveId);
      } else {
        _selectedHiveIds.remove(hiveId);
      }
    });
  }

  void _toggleCategory(_ReportCategory category) {
    setState(() {
      if (_selectedCategories.contains(category)) {
        if (_selectedCategories.length > 1) {
          _selectedCategories.remove(category);
        }
      } else {
        _selectedCategories.add(category);
      }
    });
  }

  Future<void> _pickFrom() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _range.start,
      firstDate: DateTime(DateTime.now().year - 5),
      lastDate: _range.end,
    );
    if (picked != null) {
      setState(() => _range = DateTimeRange(start: picked, end: _range.end));
    }
  }

  Future<void> _pickTo() async {
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final picked = await showDatePicker(
      context: context,
      initialDate: _range.end,
      firstDate: _range.start,
      lastDate: today,
    );
    if (picked != null) {
      setState(() => _range = DateTimeRange(start: _range.start, end: picked));
    }
  }

  bool _inRange(DateTime date) {
    final end = _range.end.add(const Duration(days: 1));
    return !date.isBefore(_range.start) && date.isBefore(end);
  }

  Future<void> _generate() async {
    if (_selectedHiveIds.isEmpty || _selectedCategories.isEmpty) return;
    setState(() {
      _loading = true;
      _report = null;
    });

    final api = context.read<ApiClient>();
    final inspectionRepo = InspectionRepository(api: api);
    final treatmentRepo = TreatmentRepository(api: api);
    final feedingRepo = FeedingRepository(api: api);
    final harvestRepo = HarvestRepository(api: api);

    final report = <int, List<_ReportEntry>>{};

    await Future.wait(_selectedHiveIds.map((hiveId) async {
      final entries = <_ReportEntry>[];

      if (_selectedCategories.contains(_ReportCategory.inspections)) {
        final result = await inspectionRepo
            .listInspections(widget.apiaryId, hiveId, limit: 200);
        for (final i in result.items) {
          if (_inRange(i.inspectedAt)) {
            entries.add(_ReportEntry(
              category: _ReportCategory.inspections,
              date: i.inspectedAt,
              record: i,
            ));
          }
        }
      }
      if (_selectedCategories.contains(_ReportCategory.treatments)) {
        final result = await treatmentRepo
            .listTreatments(widget.apiaryId, hiveId, limit: 200);
        for (final t in result.items) {
          if (_inRange(t.treatedAt)) {
            entries.add(_ReportEntry(
              category: _ReportCategory.treatments,
              date: t.treatedAt,
              record: t,
            ));
          }
        }
      }
      if (_selectedCategories.contains(_ReportCategory.feedings)) {
        final result =
            await feedingRepo.listFeedings(widget.apiaryId, hiveId, limit: 200);
        for (final f in result.items) {
          if (_inRange(f.fedAt)) {
            entries.add(_ReportEntry(
              category: _ReportCategory.feedings,
              date: f.fedAt,
              record: f,
            ));
          }
        }
      }
      if (_selectedCategories.contains(_ReportCategory.harvests)) {
        final result =
            await harvestRepo.listHarvests(widget.apiaryId, hiveId, limit: 200);
        for (final h in result.items) {
          if (_inRange(h.harvestedAt)) {
            entries.add(_ReportEntry(
              category: _ReportCategory.harvests,
              date: h.harvestedAt,
              record: h,
            ));
          }
        }
      }

      entries.sort((a, b) => b.date.compareTo(a.date));
      if (entries.isNotEmpty) report[hiveId] = entries;
    }));

    if (!mounted) return;
    setState(() {
      _loading = false;
      _report = report;
    });
  }

  Future<void> _downloadPdf() async {
    if (_selectedHiveIds.isEmpty || _selectedCategories.isEmpty) return;
    setState(() => _downloadingPdf = true);
    final dateFormat = DateFormat('yyyy-MM-dd');
    try {
      final response = await context.read<ApiClient>().dio.post<List<int>>(
        '/api/v1/apiaries/${widget.apiaryId}/report/pdf',
        data: {
          'hive_ids': _selectedHiveIds.toList(),
          'categories':
              _selectedCategories.map((c) => c.name).toList(growable: false),
          'from': dateFormat.format(_range.start),
          'to': dateFormat.format(_range.end),
        },
        options: Options(responseType: ResponseType.bytes),
      );
      final bytes = Uint8List.fromList(response.data!);
      await Printing.sharePdf(
        bytes: bytes,
        filename: 'raport-${dateFormat.format(DateTime.now())}.pdf',
      );
    } catch (_) {
      if (mounted) {
        final l10n = AppLocalizations.of(context)!;
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text(l10n.generalError)));
      }
    } finally {
      if (mounted) setState(() => _downloadingPdf = false);
    }
  }

  Future<bool> _confirmDeleteDialog(String title, String warning) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(title),
        content: Text(warning),
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
    return confirmed ?? false;
  }

  // Edit/delete actions intentionally do not auto-regenerate the report —
  // results only refresh when the user explicitly taps "Generate report"
  // again, so the on-screen cards don't shift under them mid-review.

  Future<void> _editInspection(Hive hive, Inspection inspection) async {
    await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => InspectionFormScreen(
          apiaryId: widget.apiaryId,
          hive: hive,
          inspection: inspection,
        ),
      ),
    );
  }

  Future<void> _deleteInspection(Hive hive, Inspection inspection) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await _confirmDeleteDialog(
      l10n.inspectionDeleteConfirm,
      l10n.inspectionDeleteWarning,
    );
    if (!confirmed || !mounted) return;
    await InspectionRepository(api: context.read<ApiClient>())
        .deleteInspection(
      apiaryId: widget.apiaryId,
      hiveId: hive.id,
      inspectionId: inspection.id,
    );
  }

  Future<void> _editTreatment(Hive hive, Treatment treatment) async {
    await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => TreatmentFormScreen(
          apiaryId: widget.apiaryId,
          hive: hive,
          treatment: treatment,
        ),
      ),
    );
  }

  Future<void> _deleteTreatment(Hive hive, Treatment treatment) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await _confirmDeleteDialog(
      l10n.treatmentDeleteConfirm,
      l10n.treatmentDeleteWarning,
    );
    if (!confirmed || !mounted) return;
    await TreatmentRepository(api: context.read<ApiClient>()).deleteTreatment(
      apiaryId: widget.apiaryId,
      hiveId: hive.id,
      treatmentId: treatment.id,
    );
  }

  Future<void> _editFeeding(Hive hive, Feeding feeding) async {
    await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => FeedingFormScreen(
          apiaryId: widget.apiaryId,
          hive: hive,
          feeding: feeding,
        ),
      ),
    );
  }

  Future<void> _deleteFeeding(Hive hive, Feeding feeding) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await _confirmDeleteDialog(
      l10n.feedingDeleteConfirm,
      l10n.feedingDeleteWarning,
    );
    if (!confirmed || !mounted) return;
    await FeedingRepository(api: context.read<ApiClient>()).deleteFeeding(
      apiaryId: widget.apiaryId,
      hiveId: hive.id,
      feedingId: feeding.id,
    );
  }

  Future<void> _editHarvest(Hive hive, Harvest harvest) async {
    await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => HarvestFormScreen(
          apiaryId: widget.apiaryId,
          hive: hive,
          harvest: harvest,
        ),
      ),
    );
  }

  Future<void> _deleteHarvest(Hive hive, Harvest harvest) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await _confirmDeleteDialog(
      l10n.harvestDeleteConfirm,
      l10n.harvestDeleteWarning,
    );
    if (!confirmed || !mounted) return;
    await HarvestRepository(api: context.read<ApiClient>()).deleteHarvest(
      apiaryId: widget.apiaryId,
      hiveId: hive.id,
      harvestId: harvest.id,
    );
  }

  Widget _buildEntry(Hive hive, _ReportEntry entry) {
    switch (entry.category) {
      case _ReportCategory.inspections:
        final inspection = entry.record as Inspection;
        return InspectionCard(
          apiaryId: widget.apiaryId,
          inspection: inspection,
          onEdit: () => _editInspection(hive, inspection),
          onDelete: () => _deleteInspection(hive, inspection),
        );
      case _ReportCategory.treatments:
        final treatment = entry.record as Treatment;
        return TreatmentCard(
          treatment: treatment,
          onEdit: () => _editTreatment(hive, treatment),
          onDelete: () => _deleteTreatment(hive, treatment),
        );
      case _ReportCategory.feedings:
        final feeding = entry.record as Feeding;
        return FeedingCard(
          feeding: feeding,
          onEdit: () => _editFeeding(hive, feeding),
          onDelete: () => _deleteFeeding(hive, feeding),
        );
      case _ReportCategory.harvests:
        final harvest = entry.record as Harvest;
        return HarvestCard(
          harvest: harvest,
          onEdit: () => _editHarvest(hive, harvest),
          onDelete: () => _deleteHarvest(hive, harvest),
        );
    }
  }

  Map<_ReportCategory, List<_ReportEntry>> _groupByCategory(
    List<_ReportEntry> entries,
  ) {
    final grouped = <_ReportCategory, List<_ReportEntry>>{};
    for (final category in _ReportCategory.values) {
      final matching = entries.where((e) => e.category == category).toList();
      if (matching.isNotEmpty) grouped[category] = matching;
    }
    return grouped;
  }

  String _categoryLabel(AppLocalizations l10n, _ReportCategory category) {
    return switch (category) {
      _ReportCategory.inspections => l10n.hiveDetailInspections,
      _ReportCategory.feedings => l10n.hiveDetailFeedings,
      _ReportCategory.treatments => l10n.hiveDetailTreatments,
      _ReportCategory.harvests => l10n.hiveDetailHarvests,
    };
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).toString();
    final dateFormat = DateFormat('d MMM yyyy', locale);

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.dashboardTitle),
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
                Text(l10n.dashboardHivesSection,
                    style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                if (_activeHives.isEmpty)
                  Text(l10n.hiveEmpty)
                else
                  Card(
                    child: Column(
                      children: _activeHives.map((hive) {
                        return CheckboxListTile(
                          value: _selectedHiveIds.contains(hive.id),
                          onChanged: (v) => _toggleHive(hive.id, v ?? false),
                          title: Text(hive.name),
                          secondary: _HiveStatusIcons(hive: hive),
                          controlAffinity: ListTileControlAffinity.leading,
                          dense: true,
                        );
                      }).toList(),
                    ),
                  ),
                const SizedBox(height: 20),
                Text(l10n.dashboardCategoriesSection,
                    style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: _ReportCategory.values.map((category) {
                    final selected = _selectedCategories.contains(category);
                    final isLastSelected =
                        selected && _selectedCategories.length <= 1;
                    return FilterChip(
                      label: Text(_categoryLabel(l10n, category)),
                      selected: selected,
                      onSelected: isLastSelected
                          ? null
                          : (_) => _toggleCategory(category),
                    );
                  }).toList(),
                ),
                const SizedBox(height: 20),
                Text(l10n.dashboardDateRangeSection,
                    style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                LayoutBuilder(
                  builder: (context, constraints) {
                    final fromField = _DateField(
                      label: l10n.dashboardDateFrom,
                      date: dateFormat.format(_range.start),
                      onTap: _pickFrom,
                    );
                    final toField = _DateField(
                      label: l10n.dashboardDateTo,
                      date: dateFormat.format(_range.end),
                      onTap: _pickTo,
                    );
                    // Overrides the app-wide ElevatedButton theme (which
                    // defaults to a full-width, 52dp-tall button meant for
                    // primary CTAs) so this inline button matches the
                    // height of the OutlinedButton date fields beside it.
                    final generateButton = ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        minimumSize: const Size(0, 40),
                      ),
                      onPressed: (_selectedHiveIds.isEmpty ||
                              _selectedCategories.isEmpty ||
                              _loading)
                          ? null
                          : _generate,
                      child: Text(
                        l10n.dashboardGenerate,
                        overflow: TextOverflow.ellipsis,
                      ),
                    );

                    if (constraints.maxWidth < 500) {
                      return Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Row(
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              Expanded(child: fromField),
                              const SizedBox(width: 8),
                              Expanded(child: toField),
                            ],
                          ),
                          const SizedBox(height: 12),
                          generateButton,
                        ],
                      );
                    }

                    // Date fields keep their natural content width here
                    // (not Expanded) so the group clusters together on the
                    // left instead of spreading across the full row.
                    return Row(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        fromField,
                        const SizedBox(width: 8),
                        toField,
                        const SizedBox(width: 8),
                        Flexible(child: generateButton),
                      ],
                    );
                  },
                ),
                const SizedBox(height: 24),
                if (_loading) const Center(child: CircularProgressIndicator()),
                if (!_loading && _report != null) ...[
                  if (_report!.isEmpty)
                    Center(child: Text(l10n.dashboardNoResults))
                  else
                    for (final hiveId in _selectedHiveIds)
                      if (_report![hiveId] != null) ...[
                        Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: Theme.of(context)
                                .colorScheme
                                .primary
                                .withValues(alpha: 0.06),
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(
                              color: Theme.of(context)
                                  .colorScheme
                                  .primary
                                  .withValues(alpha: 0.2),
                            ),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Icon(
                                    Icons.hive,
                                    color: Theme.of(context).colorScheme.primary,
                                  ),
                                  const SizedBox(width: 8),
                                  Text(
                                    _hiveById(hiveId).name,
                                    style: Theme.of(context)
                                        .textTheme
                                        .headlineSmall
                                        ?.copyWith(fontWeight: FontWeight.w700),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 12),
                              for (final entry in _groupByCategory(
                                _report![hiveId]!,
                              ).entries) ...[
                                Text(
                                  _categoryLabel(l10n, entry.key),
                                  style: Theme.of(context)
                                      .textTheme
                                      .titleMedium
                                      ?.copyWith(
                                        color: Theme.of(context)
                                            .colorScheme
                                            .primary,
                                        fontWeight: FontWeight.w700,
                                      ),
                                ),
                                const SizedBox(height: 8),
                                SizedBox(
                                  height: 280,
                                  child: ScrollConfiguration(
                                    behavior: _MouseDragScrollBehavior(),
                                    child: ListView.separated(
                                      scrollDirection: Axis.horizontal,
                                      itemCount: entry.value.length,
                                      separatorBuilder: (_, __) =>
                                          const SizedBox(width: 8),
                                      itemBuilder: (_, i) => SizedBox(
                                        width: 300,
                                        height: 280,
                                        child: ClipRect(
                                          child: _buildEntry(
                                            _hiveById(hiveId),
                                            entry.value[i],
                                          ),
                                        ),
                                      ),
                                    ),
                                  ),
                                ),
                                const SizedBox(height: 12),
                              ],
                            ],
                          ),
                        ),
                        const SizedBox(height: 16),
                      ],
                  if (_report!.isNotEmpty) ...[
                    const SizedBox(height: 8),
                    Center(
                      child: OutlinedButton.icon(
                        onPressed: _downloadingPdf ? null : _downloadPdf,
                        icon: _downloadingPdf
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              )
                            : const Icon(Icons.picture_as_pdf_outlined),
                        label: Text(l10n.dashboardDownloadPdf),
                      ),
                    ),
                  ],
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _DateField extends StatelessWidget {
  final String label;
  final String date;
  final VoidCallback onTap;

  const _DateField({
    required this.label,
    required this.date,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: Theme.of(context).textTheme.labelSmall),
        const SizedBox(height: 4),
        OutlinedButton(
          onPressed: onTap,
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.date_range_outlined, size: 18),
              const SizedBox(width: 6),
              Text(date, overflow: TextOverflow.ellipsis),
            ],
          ),
        ),
      ],
    );
  }
}

class _HiveStatusIcons extends StatelessWidget {
  final Hive hive;

  const _HiveStatusIcons({required this.hive});

  @override
  Widget build(BuildContext context) {
    final icons = [
      if (hive.queenless) Icons.female_outlined,
      if (hive.readyForHarvest) Icons.water_drop_outlined,
      if (hive.needsFood) Icons.restaurant_outlined,
      if (hive.diseases.isNotEmpty) Icons.coronavirus_outlined,
    ];
    if (icons.isEmpty) return const SizedBox.shrink();
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: icons
          .map((icon) => Padding(
                padding: const EdgeInsets.only(left: 4),
                child: Icon(icon, size: 18),
              ))
          .toList(),
    );
  }
}

/// Allows the horizontal report card list to be scrolled by clicking and
/// dragging with a mouse — Flutter's default desktop/web scroll behavior
/// only enables drag-to-scroll for touch and trackpad, not mouse.
class _MouseDragScrollBehavior extends MaterialScrollBehavior {
  @override
  Set<PointerDeviceKind> get dragDevices => {
        ...super.dragDevices,
        PointerDeviceKind.mouse,
      };
}
