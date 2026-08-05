import 'dart:async';
import 'dart:io' show File;
import 'dart:math';

import 'package:audioplayers/audioplayers.dart';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';
import 'package:path_provider/path_provider.dart';
import 'package:record/record.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/delete_dialog.dart';
import '../../../core/widgets/photo_size_snackbar.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../../apiary/data/apiary_model.dart';
import '../../dashboard/view/dashboard_screen.dart';
import '../../feeding/data/feeding_model.dart';
import '../../feeding/data/feeding_repository.dart';
import '../../feeding/view/feeding_form_screen.dart';
import '../../feeding/view/feeding_summary.dart';
import '../../harvest/data/harvest_model.dart';
import '../../harvest/data/harvest_repository.dart';
import '../../harvest/view/harvest_form_screen.dart';
import '../../harvest/view/harvest_summary.dart';
import '../../inspection/data/inspection_model.dart';
import '../../inspection/data/inspection_repository.dart';
import '../../inspection/view/inspection_form_screen.dart';
import '../../inspection/view/inspection_summary.dart';
import '../../treatment/data/treatment_model.dart';
import '../../treatment/data/treatment_repository.dart';
import '../../treatment/view/treatment_form_screen.dart';
import '../../treatment/view/treatment_summary.dart';
import '../../voice/data/voice_recording_model.dart';
import '../../voice/data/voice_repository.dart';
import '../cubit/hives_cubit.dart';
import '../data/hive_model.dart';
import '../data/hive_repository.dart';
import '../../treatment/view/bulk_treatment_form_screen.dart';
import '../../feeding/view/bulk_feeding_form_screen.dart';
import 'add_hive_screen.dart';
import 'bulk_hive_selection_dialog.dart';
import 'edit_hive_screen.dart';
import 'hive_detail_screen.dart';
import 'hive_form_widgets.dart';
import 'update_hive_status_form_screen.dart';

enum _HiveFilter {
  readyForHarvest,
  queenNeedsReplacement,
  needsFood,
  boxNeedsAdding,
  sick,
}

class ApiaryGridScreen extends StatelessWidget {
  final Apiary apiary;

  const ApiaryGridScreen({super.key, required this.apiary});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => HivesCubit(
        repo: HiveRepository(api: context.read()),
        apiaryId: apiary.id,
      )..load(),
      child: _ApiaryGridView(apiary: apiary),
    );
  }
}

class _ApiaryGridView extends StatefulWidget {
  final Apiary apiary;

  const _ApiaryGridView({required this.apiary});

  @override
  State<_ApiaryGridView> createState() => _ApiaryGridViewState();
}

class _ApiaryGridViewState extends State<_ApiaryGridView> {
  final Set<_HiveFilter> _activeFilters = {};
  final TransformationController _transformController =
      TransformationController();
  int _pendingReviewCount = 0;

  @override
  void initState() {
    super.initState();
    _loadPendingReviewCount();
  }

  @override
  void dispose() {
    _transformController.dispose();
    super.dispose();
  }

  Future<void> _loadPendingReviewCount() async {
    try {
      final repo = VoiceRepository(api: context.read<ApiClient>());
      final result = await repo.listRecordings(widget.apiary.id);
      final count = result.items
          .where((r) => r.status == voiceRecordingStatusCompleted)
          .length;
      if (mounted) setState(() => _pendingReviewCount = count);
    } catch (_) {
      // Best-effort — the badge just stays at its last known value.
    }
  }

  void _toggleFilter(_HiveFilter filter) {
    setState(() {
      if (_activeFilters.contains(filter)) {
        _activeFilters.remove(filter);
      } else {
        _activeFilters.add(filter);
      }
    });
  }

  bool _matches(Hive hive) {
    if (_activeFilters.isEmpty) return true;
    if (_activeFilters.contains(_HiveFilter.readyForHarvest) &&
        hive.readyForHarvest) return true;
    if (_activeFilters.contains(_HiveFilter.queenNeedsReplacement) &&
        hive.queenNeedsReplacement) {
      return true;
    }
    if (_activeFilters.contains(_HiveFilter.needsFood) && hive.needsFood) {
      return true;
    }
    if (_activeFilters.contains(_HiveFilter.boxNeedsAdding) &&
        hive.boxNeedsAdding) {
      return true;
    }
    if (_activeFilters.contains(_HiveFilter.sick) && hive.diseases.isNotEmpty) {
      return true;
    }
    return false;
  }

  Future<void> _openDetail(BuildContext context, Hive hive) async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => HiveDetailScreen(hive: hive, apiaryId: widget.apiary.id),
      ),
    );
    if (context.mounted) context.read<HivesCubit>().load();
  }

  Future<void> _openBulkTreatment(BuildContext context, List<Hive> hives) async {
    final activeHives = hives.where((h) => h.active).toList();
    final hiveIds =
        await showBulkHiveSelectionDialog(context, hives: activeHives);
    if (hiveIds == null || !context.mounted) return;
    final l10n = AppLocalizations.of(context)!;
    final count = await Navigator.of(context).push<int>(
      MaterialPageRoute(
        builder: (_) => BulkTreatmentFormScreen(
          apiaryId: widget.apiary.id,
          hiveIds: hiveIds,
        ),
      ),
    );
    if (count != null && context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.treatmentBulkSuccess(count))),
      );
    }
  }

  Future<void> _openBulkFeeding(BuildContext context, List<Hive> hives) async {
    final activeHives = hives.where((h) => h.active).toList();
    final hiveIds =
        await showBulkHiveSelectionDialog(context, hives: activeHives);
    if (hiveIds == null || !context.mounted) return;
    final l10n = AppLocalizations.of(context)!;
    final count = await Navigator.of(context).push<int>(
      MaterialPageRoute(
        builder: (_) => BulkFeedingFormScreen(
          apiaryId: widget.apiary.id,
          hiveIds: hiveIds,
        ),
      ),
    );
    if (count != null && context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.feedingBulkSuccess(count))),
      );
    }
  }

  Future<void> _openAddHive(
    BuildContext context,
    HivesLoaded state,
    int row,
    int col,
  ) async {
    final l10n = AppLocalizations.of(context)!;
    final defaultName = l10n.hiveDefaultName(state.hives.length + 1);
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => AddHiveScreen(
          apiaryId: widget.apiary.id,
          gridRow: row,
          gridCol: col,
          defaultName: defaultName,
          existingNames:
              state.hives.map((h) => h.name.trim().toLowerCase()).toSet(),
        ),
      ),
    );
    if (context.mounted) context.read<HivesCubit>().load();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(widget.apiary.name),
        actions: const [ProfileIconButton()],
      ),
      body: BlocBuilder<HivesCubit, HivesState>(
        builder: (context, state) {
          if (state is HivesInitial || state is HivesLoading) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is HivesError) {
            return Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    l10n.generalError,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 12),
                  ElevatedButton(
                    onPressed: () => context.read<HivesCubit>().load(),
                    style: ElevatedButton.styleFrom(
                      minimumSize: const Size(120, 40),
                    ),
                    child: Text(l10n.generalRetry),
                  ),
                ],
              ),
            );
          }
          if (state is HivesLoaded) {
            final hiveMap = {
              for (final h in state.hives) (h.gridRow, h.gridCol): h,
            };
            const double cellSize = 80;
            const double spacing = 8;
            const double padding = 16;
            final gridContentWidth = widget.apiary.gridCols * cellSize +
                (widget.apiary.gridCols - 1) * spacing;
            return Column(
              children: [
                Expanded(
                  child: LayoutBuilder(
                    builder: (context, constraints) {
                      final gridContentHeight =
                          widget.apiary.gridRows * cellSize +
                              (widget.apiary.gridRows - 1) * spacing;
                      final contentW = gridContentWidth + padding * 2;
                      final contentH = gridContentHeight + padding * 2;
                      final canvasW =
                          max(constraints.maxWidth, contentW);
                      final canvasH =
                          max(constraints.maxHeight, contentH);
                      return InteractiveViewer(
                        transformationController: _transformController,
                        constrained: false,
                        boundaryMargin:
                            const EdgeInsets.all(double.infinity),
                        minScale: 0.3,
                        maxScale: 4.0,
                        trackpadScrollCausesScale: true,
                        child: SizedBox(
                          width: canvasW,
                          height: canvasH,
                          child: Center(
                            child: Padding(
                              padding: const EdgeInsets.all(padding),
                              child: SizedBox(
                                width: gridContentWidth,
                                child: GridView.builder(
                                  shrinkWrap: true,
                                  physics:
                                      const NeverScrollableScrollPhysics(),
                                  gridDelegate:
                                      SliverGridDelegateWithFixedCrossAxisCount(
                                    crossAxisCount: widget.apiary.gridCols,
                                    crossAxisSpacing: spacing,
                                    mainAxisSpacing: spacing,
                                  ),
                                  itemCount: widget.apiary.gridRows *
                                      widget.apiary.gridCols,
                                  itemBuilder: (context, index) {
                                    final row =
                                        index ~/ widget.apiary.gridCols;
                                    final col =
                                        index % widget.apiary.gridCols;
                                    final hive = hiveMap[(row, col)];
                                    return hive != null
                                        ? _HiveCell(
                                            hive: hive,
                                            dimmed: _activeFilters
                                                    .isNotEmpty &&
                                                !_matches(hive),
                                            onTap: () => _openDetail(
                                                context, hive),
                                          )
                                        : _EmptyCell(
                                            onTap: () => _openAddHive(
                                                context, state, row, col),
                                            onDrop: (h) => context
                                                .read<HivesCubit>()
                                                .move(h.id, row, col),
                                          );
                                  },
                                ),
                              ),
                            ),
                          ),
                        ),
                      );
                    },
                  ),
                ),
                _FilterBar(
                  activeFilters: _activeFilters,
                  onToggle: _toggleFilter,
                  onCenter: () =>
                      _transformController.value = Matrix4.identity(),
                  onTreatAll: () => _openBulkTreatment(context, state.hives),
                  onFeedAll: () => _openBulkFeeding(context, state.hives),
                  pendingReviewCount: _pendingReviewCount,
                  onVoiceDialogClosed: _loadPendingReviewCount,
                  onDashboard: () => Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (_) => DashboardScreen(
                        apiaryId: widget.apiary.id,
                        hives: state.hives,
                      ),
                    ),
                  ),
                  l10n: l10n,
                  apiaryId: widget.apiary.id,
                  hives: state.hives,
                  onHiveTap: (hive) async {
                    await Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => HiveDetailScreen(
                          hive: hive,
                          apiaryId: widget.apiary.id,
                        ),
                      ),
                    );
                    if (context.mounted) context.read<HivesCubit>().load();
                  },
                ),
              ],
            );
          }
          return const SizedBox.shrink();
        },
      ),
    );
  }
}

class _FilterBar extends StatelessWidget {
  final Set<_HiveFilter> activeFilters;
  final void Function(_HiveFilter) onToggle;
  final VoidCallback onCenter;
  final VoidCallback onTreatAll;
  final VoidCallback onFeedAll;
  final VoidCallback onDashboard;
  final AppLocalizations l10n;
  final int apiaryId;
  final List<Hive> hives;
  final void Function(Hive) onHiveTap;
  final int pendingReviewCount;
  final VoidCallback onVoiceDialogClosed;

  const _FilterBar({
    required this.activeFilters,
    required this.onToggle,
    required this.onCenter,
    required this.onTreatAll,
    required this.onFeedAll,
    required this.onDashboard,
    required this.l10n,
    required this.apiaryId,
    required this.hives,
    required this.onHiveTap,
    required this.pendingReviewCount,
    required this.onVoiceDialogClosed,
  });

  void _showFilterSheet(BuildContext context) {
    final localFilters = Set<_HiveFilter>.from(activeFilters);
    final chips = [
      (_HiveFilter.readyForHarvest, l10n.hiveReadyForHarvest),
      (_HiveFilter.queenNeedsReplacement, l10n.hiveQueenNeedsReplacement),
      (_HiveFilter.needsFood, l10n.hiveNeedsFood),
      (_HiveFilter.boxNeedsAdding, l10n.hiveBoxNeedsAdding),
      (_HiveFilter.sick, l10n.hiveSick),
    ];
    showDialog<void>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          final isWide = MediaQuery.sizeOf(ctx).width >= 600;
          return Dialog(
            shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(20)),
            child: ConstrainedBox(
              constraints: BoxConstraints(maxWidth: isWide ? 480 : 360),
              child: Padding(
                padding:
                    const EdgeInsets.symmetric(horizontal: 24, vertical: 24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Row(
                      children: [
                        Icon(Icons.tune,
                            size: 20,
                            color: Theme.of(ctx)
                                .colorScheme
                                .onSurfaceVariant),
                        const SizedBox(width: 8),
                        Text(l10n.hiveFilterTooltip,
                            style: Theme.of(ctx).textTheme.titleMedium),
                        const Spacer(),
                        IconButton(
                          icon: const Icon(Icons.close),
                          visualDensity: VisualDensity.compact,
                          onPressed: () => Navigator.of(ctx).pop(),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Wrap(
                      alignment: WrapAlignment.center,
                      spacing: 10,
                      runSpacing: 8,
                      children: chips.map((entry) {
                        final (filter, label) = entry;
                        final selected = localFilters.contains(filter);
                        return FilterChip(
                          label: Text(label),
                          padding: const EdgeInsets.symmetric(
                              horizontal: 12, vertical: 8),
                          selected: selected,
                          onSelected: (_) {
                            onToggle(filter);
                            setDialogState(() {
                              if (localFilters.contains(filter)) {
                                localFilters.remove(filter);
                              } else {
                                localFilters.add(filter);
                              }
                            });
                          },
                        );
                      }).toList(),
                    ),
                  ],
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  void _showHiveSheet(BuildContext context) {
    showDialog<void>(
      context: context,
      builder: (_) => _HiveListDialog(
        apiaryId: apiaryId,
        hives: hives,
        onHiveTap: onHiveTap,
        onTreatAll: onTreatAll,
        onFeedAll: onFeedAll,
      ),
    );
  }

  Future<void> _showVoiceRecordingSheet(BuildContext context) async {
    await showDialog<void>(
      context: context,
      builder: (_) => _VoiceRecordingDialog(
        apiaryId: apiaryId,
        onRecordingResolved: onVoiceDialogClosed,
      ),
    );
    onVoiceDialogClosed();
  }

  @override
  Widget build(BuildContext context) {
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
                  if (hives.isNotEmpty)
                    Badge(
                      isLabelVisible: activeFilters.isNotEmpty,
                      label: Text('${activeFilters.length}'),
                      child: IconButton(
                        icon: const Icon(Icons.tune),
                        iconSize: 28,
                        tooltip: l10n.hiveFilterTooltip,
                        onPressed: () => _showFilterSheet(context),
                      ),
                    ),
                  if (hives.isNotEmpty)
                    IconButton(
                      icon: const Icon(Icons.format_list_bulleted),
                      iconSize: 28,
                      tooltip: l10n.hiveListTooltip,
                      onPressed: () => _showHiveSheet(context),
                    ),
                  if (hives.isNotEmpty)
                    Badge(
                      backgroundColor: Colors.red,
                      isLabelVisible: pendingReviewCount > 0,
                      label: Text('$pendingReviewCount'),
                      child: IconButton(
                        icon: const Icon(Icons.mic_none),
                        iconSize: 28,
                        tooltip: l10n.voiceRecordingTooltip,
                        onPressed: () => _showVoiceRecordingSheet(context),
                      ),
                    ),
                  IconButton(
                    icon: const Icon(Icons.center_focus_strong_outlined),
                    iconSize: 28,
                    tooltip: l10n.apiaryCenterView,
                    onPressed: onCenter,
                  ),
                  if (hives.isNotEmpty)
                    IconButton(
                      icon: const Icon(Icons.assessment_outlined),
                      iconSize: 28,
                      tooltip: l10n.dashboardTitle,
                      onPressed: onDashboard,
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

class _HiveListDialog extends StatefulWidget {
  final int apiaryId;
  final List<Hive> hives;
  final void Function(Hive) onHiveTap;
  final VoidCallback onTreatAll;
  final VoidCallback onFeedAll;

  const _HiveListDialog({
    required this.apiaryId,
    required this.hives,
    required this.onHiveTap,
    required this.onTreatAll,
    required this.onFeedAll,
  });

  @override
  State<_HiveListDialog> createState() => _HiveListDialogState();
}

class _HiveListDialogState extends State<_HiveListDialog> {
  final Map<int, Inspection?> _lastInspections = {};

  @override
  void initState() {
    super.initState();
    _loadInspections();
  }

  Future<void> _loadInspections() async {
    final repo = InspectionRepository(api: context.read<ApiClient>());
    final hivesWithInspections =
        widget.hives.where((h) => h.lastInspectedAt != null).toList();
    await Future.wait(
      hivesWithInspections.map((hive) async {
        try {
          final list = await repo.listInspections(
            widget.apiaryId,
            hive.id,
            limit: 1,
          );
          if (mounted) {
            setState(() {
              _lastInspections[hive.id] = list.items.isNotEmpty ? list.items.first : null;
            });
          }
        } catch (_) {}
      }),
    );
  }

  String _hint(AppLocalizations l10n, Inspection inspection) {
    if (inspection.notes.isNotEmpty) return inspection.notes;
    final parts = <String>[];
    if (inspection.queenSeen == 'seen') {
      parts.add(l10n.inspectionQueenStatusSeen);
    } else if (inspection.queenSeen == 'not_seen') {
      parts.add(l10n.inspectionQueenStatusNotSeen);
    }
    if (inspection.broodPattern.isNotEmpty &&
        inspection.broodPattern != 'none') {
      final broodLabel = switch (inspection.broodPattern) {
        'excellent' => l10n.inspectionBroodExcellent,
        'good' => l10n.inspectionBroodGood,
        'poor' => l10n.inspectionBroodPoor,
        _ => inspection.broodPattern,
      };
      parts.add('${l10n.inspectionBroodPattern}: $broodLabel');
    }
    return parts.join(' · ');
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).toString();
    final isWide = MediaQuery.sizeOf(context).width >= 600;

    final sortedHives = [...widget.hives]..sort((a, b) {
        if (a.lastInspectedAt == null && b.lastInspectedAt == null) return 0;
        if (a.lastInspectedAt == null) return -1;
        if (b.lastInspectedAt == null) return 1;
        return a.lastInspectedAt!.compareTo(b.lastInspectedAt!);
      });

    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: isWide ? 480 : 380),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(24, 24, 24, 8),
              child: Row(
                children: [
                  Icon(Icons.format_list_bulleted,
                      size: 20,
                      color: Theme.of(context).colorScheme.onSurfaceVariant),
                  const SizedBox(width: 8),
                  Text(l10n.hiveListTooltip,
                      style: Theme.of(context).textTheme.titleMedium),
                  const Spacer(),
                  IconButton(
                    icon: const Icon(Icons.close),
                    visualDensity: VisualDensity.compact,
                    onPressed: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
            ),
            const Divider(height: 1),
            Flexible(
              child: ListView.builder(
                shrinkWrap: true,
                itemCount: sortedHives.length,
                itemBuilder: (_, i) {
                  final hive = sortedHives[i];
                  final dateStr = hive.lastInspectedAt != null
                      ? DateFormat('d MMM yyyy', locale)
                          .format(hive.lastInspectedAt!)
                      : l10n.hiveDetailNoInspections;

                  final inspection = _lastInspections[hive.id];
                  String subtitleText = dateStr;
                  if (isWide &&
                      hive.lastInspectedAt != null &&
                      inspection != null) {
                    final hint = _hint(l10n, inspection);
                    if (hint.isNotEmpty) subtitleText = '$dateStr · $hint';
                  }

                  return ListTile(
                    leading: const Icon(Icons.hive),
                    title: Text(hive.name),
                    subtitle: Text(
                      subtitleText,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    trailing: _HiveStatusIcons(hive: hive),
                    onTap: () {
                      Navigator.of(context).pop();
                      widget.onHiveTap(hive);
                    },
                  );
                },
              ),
            ),
            if (widget.hives.length > 1) ...[
              const Divider(height: 1),
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 12),
                child: Center(
                  child: Wrap(
                    alignment: WrapAlignment.center,
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      OutlinedButton.icon(
                        onPressed: () {
                          Navigator.of(context).pop();
                          widget.onTreatAll();
                        },
                        icon: const Icon(Icons.medical_services_outlined, size: 18),
                        label: Text(l10n.treatmentTreatAllHives),
                      ),
                      OutlinedButton.icon(
                        onPressed: () {
                          Navigator.of(context).pop();
                          widget.onFeedAll();
                        },
                        icon: const Icon(Icons.grain, size: 18),
                        label: Text(l10n.feedingFeedAllHives),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

const _voiceHardCapDuration = Duration(minutes: 3);
const _voiceHardCapWarningThreshold = Duration(seconds: 10);
const _voiceSilenceAutoStopDuration = Duration(milliseconds: 2500);
const _voiceSilenceThresholdDb = -35.0;
const _voiceAmplitudePollInterval = Duration(milliseconds: 300);
const _maxClientPendingRecordings = 3;

/// In-memory cache of this session's locally-recorded pending/processing
/// voice recordings, keyed by apiary id. Survives the recording dialog
/// being closed and reopened (but not an app restart) — audio is only ever
/// playable from this device's local copy, so there's no server-side
/// source of truth to refetch it from once the dialog widget is disposed.
class _VoicePendingCache {
  static final Map<int, List<VoiceRecording>> _byApiary = {};

  static List<VoiceRecording> forApiary(int apiaryId) =>
      List.unmodifiable(_byApiary[apiaryId] ?? const []);

  static void setForApiary(int apiaryId, List<VoiceRecording> recordings) {
    _byApiary[apiaryId] = recordings;
  }
}

class _VoiceRecordingDialog extends StatefulWidget {
  final int apiaryId;
  final VoidCallback onRecordingResolved;

  const _VoiceRecordingDialog({
    required this.apiaryId,
    required this.onRecordingResolved,
  });

  @override
  State<_VoiceRecordingDialog> createState() => _VoiceRecordingDialogState();
}

class _VoiceRecordingDialogState extends State<_VoiceRecordingDialog> {
  final AudioRecorder _recorder = AudioRecorder();
  final AudioPlayer _player = AudioPlayer();
  bool _isRecording = false;
  bool _hasDetectedSpeech = false;
  bool _hardCapWarning = false;
  bool _uploading = false;
  Duration _elapsed = Duration.zero;
  DateTime? _lastSpeechAt;
  StreamSubscription<Amplitude>? _ampSub;
  Timer? _tickTimer;
  Timer? _pollTimer;
  StreamSubscription<void>? _playbackCompleteSub;
  String? _playingPath;
  late List<VoiceRecording> _pending;
  List<VoiceRecording> _readyForReview = [];
  final ScrollController _pendingScrollController = ScrollController();
  final ScrollController _readyForReviewScrollController = ScrollController();

  bool get _atPendingLimit => _pending.length >= _maxClientPendingRecordings;

  @override
  void initState() {
    super.initState();
    _pending = _VoicePendingCache.forApiary(widget.apiaryId);
    if (_pending.isNotEmpty) _startPolling();
    _playbackCompleteSub = _player.onPlayerComplete.listen((_) {
      if (mounted) setState(() => _playingPath = null);
    });
    _refreshFromServer();
  }

  /// Reconciles the locally-cached [_pending] list against the server's
  /// current state and (re)loads [_readyForReview] from the same fetch, so
  /// a recording can never show as both — e.g. a recording that finished
  /// (or failed) while the dialog was closed no longer lingers as "still
  /// pending" from the stale session cache once this returns.
  Future<void> _refreshFromServer() async {
    try {
      final repo = VoiceRepository(api: context.read<ApiClient>());
      final result = await repo.listRecordings(widget.apiaryId);
      final byId = {for (final r in result.items) r.recordingId: r};
      final stillPending = <VoiceRecording>[];
      for (final rec in _pending) {
        final status = byId[rec.recordingId]?.status;
        if (status == voiceRecordingStatusPending ||
            status == voiceRecordingStatusProcessing) {
          stillPending.add(rec.copyWith(status: status));
        } else {
          await _discardLocalFile(rec.localPath);
          if (rec.localPath != null) {
            await _stopPlaybackIfPlaying(rec.localPath!);
          }
        }
      }
      final ready = result.items
          .where((r) => r.status == voiceRecordingStatusCompleted)
          .toList();
      if (!mounted) return;
      setState(() {
        _pending = stillPending;
        _readyForReview = ready;
      });
      _VoicePendingCache.setForApiary(widget.apiaryId, _pending);
    } catch (_) {
      // Best-effort — both sections just stay at their last known value.
    }
  }

  Future<void> _toggleRecording() async {
    if (_isRecording) {
      await _stopRecording();
    } else {
      await _startRecording();
    }
  }

  Future<String> _recordingPath() async {
    final ext = kIsWeb ? 'webm' : 'm4a';
    final filename = 'voice_${DateTime.now().millisecondsSinceEpoch}.$ext';
    if (kIsWeb) return filename;
    final dir = await getTemporaryDirectory();
    return '${dir.path}/$filename';
  }

  Future<void> _startRecording() async {
    if (_atPendingLimit) return;

    final hasPermission = await _recorder.hasPermission();
    if (!mounted) return;
    if (!hasPermission) {
      showBigSnackBar(
          context, AppLocalizations.of(context)!.voiceRecordingPermissionDenied);
      return;
    }

    final path = await _recordingPath();
    final encoder = kIsWeb ? AudioEncoder.opus : AudioEncoder.aacLc;
    await _recorder.start(RecordConfig(encoder: encoder), path: path);
    if (!mounted) return;

    setState(() {
      _isRecording = true;
      _hasDetectedSpeech = false;
      _hardCapWarning = false;
      _elapsed = Duration.zero;
      _lastSpeechAt = null;
    });

    _ampSub = _recorder
        .onAmplitudeChanged(_voiceAmplitudePollInterval)
        .listen(_onAmplitude);
    _tickTimer = Timer.periodic(const Duration(seconds: 1), (_) => _onTick());
  }

  void _onAmplitude(Amplitude amp) {
    if (amp.current > _voiceSilenceThresholdDb) {
      _hasDetectedSpeech = true;
      _lastSpeechAt = DateTime.now();
      return;
    }
    if (!_hasDetectedSpeech || _lastSpeechAt == null) return;
    if (DateTime.now().difference(_lastSpeechAt!) >=
        _voiceSilenceAutoStopDuration) {
      _stopRecording();
    }
  }

  void _onTick() {
    if (!_isRecording) return;
    setState(() {
      _elapsed += const Duration(seconds: 1);
      _hardCapWarning =
          _voiceHardCapDuration - _elapsed <= _voiceHardCapWarningThreshold;
    });
    if (_elapsed >= _voiceHardCapDuration) {
      _stopRecording();
    }
  }

  Future<void> _stopRecording() async {
    if (!_isRecording) return;
    await _ampSub?.cancel();
    _ampSub = null;
    _tickTimer?.cancel();
    _tickTimer = null;

    final path = await _recorder.stop();

    if (!mounted) return;
    setState(() {
      _isRecording = false;
      _hasDetectedSpeech = false;
      _hardCapWarning = false;
      _elapsed = Duration.zero;
    });

    if (path != null) await _uploadRecording(path);
  }

  Future<List<int>> _fetchBlobBytes(String blobUrl) async {
    final response = await Dio().get<List<int>>(
      blobUrl,
      options: Options(responseType: ResponseType.bytes),
    );
    return response.data!;
  }

  Future<void> _uploadRecording(String path) async {
    final repo = VoiceRepository(api: context.read<ApiClient>());
    setState(() => _uploading = true);
    try {
      final bytes = kIsWeb
          ? await _fetchBlobBytes(path)
          : await File(path).readAsBytes();
      final uploaded = await repo.uploadRecording(
        widget.apiaryId,
        bytes,
        filename: kIsWeb ? 'voice.webm' : 'voice.m4a',
        mimeType: kIsWeb ? 'audio/webm' : 'audio/mp4',
      );
      if (!mounted) return;
      _pending = [uploaded.copyWith(localPath: path), ..._pending];
      _VoicePendingCache.setForApiary(widget.apiaryId, _pending);
      setState(() {});
      _startPolling();
    } catch (_) {
      await _discardLocalFile(path);
      if (mounted) {
        showBigSnackBar(
            context, AppLocalizations.of(context)!.voiceRecordingUploadFailed);
      }
    } finally {
      if (mounted) setState(() => _uploading = false);
    }
  }

  void _startPolling() {
    _pollTimer?.cancel();
    _pollTimer =
        Timer.periodic(const Duration(seconds: 4), (_) => _pollPending());
  }

  Future<void> _pollPending() async {
    if (_pending.isEmpty) {
      _pollTimer?.cancel();
      _pollTimer = null;
      return;
    }
    try {
      final repo = VoiceRepository(api: context.read<ApiClient>());
      final result = await repo.listRecordings(widget.apiaryId);
      final byId = {for (final r in result.items) r.recordingId: r};
      final stillPending = <VoiceRecording>[];
      final newlyReady = <VoiceRecording>[];
      for (final rec in _pending) {
        final updated = byId[rec.recordingId];
        final status = updated?.status;
        if (status == voiceRecordingStatusPending ||
            status == voiceRecordingStatusProcessing) {
          stillPending.add(rec.copyWith(status: status));
        } else {
          await _discardLocalFile(rec.localPath);
          if (rec.localPath != null) {
            await _stopPlaybackIfPlaying(rec.localPath!);
          }
          if (status == voiceRecordingStatusCompleted && updated != null) {
            newlyReady.add(updated);
          }
        }
      }
      if (!mounted) return;
      setState(() {
        _pending = stillPending;
        if (newlyReady.isNotEmpty) {
          final existingIds =
              _readyForReview.map((r) => r.recordingId).toSet();
          final trulyNew = newlyReady
              .where((r) => !existingIds.contains(r.recordingId))
              .toList();
          _readyForReview = [...trulyNew, ..._readyForReview];
        }
      });
      _VoicePendingCache.setForApiary(widget.apiaryId, _pending);
      if (_pending.isEmpty) {
        _pollTimer?.cancel();
        _pollTimer = null;
      }
    } catch (_) {
      // Best-effort — try again on the next tick.
    }
  }

  Future<void> _discardLocalFile(String? path) async {
    if (path == null || kIsWeb) return;
    final file = File(path);
    if (await file.exists()) await file.delete();
  }

  Future<void> _playRecording(String path) async {
    if (_playingPath == path) {
      await _player.stop();
      if (mounted) setState(() => _playingPath = null);
      return;
    }

    await _player.stop();
    setState(() => _playingPath = path);
    await _player.play(kIsWeb ? UrlSource(path) : DeviceFileSource(path));
  }

  Future<void> _stopPlaybackIfPlaying(String path) async {
    if (_playingPath != path) return;
    await _player.stop();
    if (mounted) setState(() => _playingPath = null);
  }

  Future<void> _cancelRecording(VoiceRecording recording) async {
    try {
      final repo = VoiceRepository(api: context.read<ApiClient>());
      await repo.cancelRecording(widget.apiaryId, recording.recordingId);
      await _discardLocalFile(recording.localPath);
      if (recording.localPath != null) {
        await _stopPlaybackIfPlaying(recording.localPath!);
      }
      if (!mounted) return;
      setState(() {
        _pending = _pending
            .where((r) => r.recordingId != recording.recordingId)
            .toList();
      });
      _VoicePendingCache.setForApiary(widget.apiaryId, _pending);
    } catch (_) {
      if (mounted) {
        showBigSnackBar(
            context, AppLocalizations.of(context)!.voiceRecordingCancelFailed);
      }
    }
  }

  void _showRecordingDetail(VoiceRecording recording) {
    showDialog<void>(
      context: context,
      builder: (_) => _VoiceRecordingDetailDialog(
        apiaryId: widget.apiaryId,
        recording: recording,
        onAccept: () => _acceptRecording(recording),
        onReject: () => _rejectRecording(recording),
      ),
    );
  }

  void _removeFromReadyForReview(VoiceRecording recording) {
    setState(() {
      _readyForReview = _readyForReview
          .where((r) => r.recordingId != recording.recordingId)
          .toList();
    });
    widget.onRecordingResolved();
  }

  bool _hasResolvedHive(VoiceRecording recording) =>
      recording.voiceActions.any((a) => a.hiveId != null);

  Future<bool> _rejectRecording(VoiceRecording recording) async {
    final l10n = AppLocalizations.of(context)!;
    if (_hasResolvedHive(recording)) {
      final confirmed = await showDeleteDialog(
        context,
        title: l10n.voiceReviewRejectConfirmTitle,
        warning: l10n.voiceReviewRejectConfirmWarning,
        l10n: l10n,
        withPuzzle: true,
        confirmLabel: l10n.voiceReviewRejectAction,
      );
      if (!confirmed || !mounted) return false;
    }
    try {
      final repo = VoiceRepository(api: context.read<ApiClient>());
      await repo.rejectRecording(widget.apiaryId, recording.recordingId);
      if (!mounted) return false;
      _removeFromReadyForReview(recording);
      return true;
    } catch (_) {
      if (mounted) {
        showBigSnackBar(
            context, AppLocalizations.of(context)!.voiceReviewRejectFailed);
      }
      return false;
    }
  }

  Future<bool> _acceptRecording(VoiceRecording recording) async {
    try {
      final repo = VoiceRepository(api: context.read<ApiClient>());
      await repo.acceptRecording(widget.apiaryId, recording.recordingId);
      if (!mounted) return false;
      _removeFromReadyForReview(recording);
      return true;
    } catch (_) {
      if (mounted) {
        showBigSnackBar(
            context, AppLocalizations.of(context)!.voiceReviewAcceptFailed);
      }
      return false;
    }
  }

  String _formatElapsed(Duration d) {
    final minutes = d.inMinutes.toString().padLeft(2, '0');
    final seconds = (d.inSeconds % 60).toString().padLeft(2, '0');
    return '$minutes:$seconds';
  }

  @override
  void dispose() {
    _ampSub?.cancel();
    _tickTimer?.cancel();
    _pollTimer?.cancel();
    _playbackCompleteSub?.cancel();
    _pendingScrollController.dispose();
    _readyForReviewScrollController.dispose();
    if (_isRecording) {
      _recorder.stop();
    }
    _recorder.dispose();
    _player.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final screenSize = MediaQuery.sizeOf(context);
    final isWide = screenSize.width >= 600;
    final remainingSeconds =
        (_voiceHardCapDuration - _elapsed).inSeconds.clamp(0, 999);
    final listMaxHeight = (screenSize.height * 0.3).clamp(200.0, 340.0);

    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: isWide ? 560 : 440),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(24, 24, 24, 32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(
                children: [
                  Icon(Icons.mic_none,
                      size: 20, color: colorScheme.onSurfaceVariant),
                  const SizedBox(width: 8),
                  Text(l10n.voiceRecordingTooltip,
                      style: Theme.of(context).textTheme.titleMedium),
                  const Spacer(),
                  IconButton(
                    icon: const Icon(Icons.close),
                    visualDensity: VisualDensity.compact,
                    onPressed: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
              const Divider(height: 1),
              const SizedBox(height: 24),
              GestureDetector(
                onTap: _atPendingLimit && !_isRecording
                    ? null
                    : _toggleRecording,
                child: CircleAvatar(
                  radius: 40,
                  backgroundColor: _isRecording
                      ? (_hardCapWarning ? Colors.deepOrange : Colors.red)
                      : _atPendingLimit
                          ? colorScheme.primary.withAlpha(90)
                          : colorScheme.primary,
                  child: Icon(
                    _isRecording ? Icons.stop : Icons.mic,
                    size: 36,
                    color: colorScheme.onPrimary,
                  ),
                ),
              ),
              const SizedBox(height: 16),
              if (_isRecording) ...[
                Text(
                  _formatElapsed(_elapsed),
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 4),
              ],
              Text(
                !_isRecording
                    ? (_atPendingLimit
                        ? l10n.voiceRecordingPendingLimitReached(
                            _maxClientPendingRecordings)
                        : l10n.voiceRecordingHintIdle)
                    : _hardCapWarning
                        ? l10n.voiceRecordingHardCapWarning(remainingSeconds)
                        : l10n.voiceRecordingHintActive,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: _hardCapWarning
                          ? Colors.deepOrange
                          : colorScheme.onSurfaceVariant,
                    ),
                textAlign: TextAlign.center,
              ),
              if (_uploading) ...[
                const SizedBox(height: 16),
                const SizedBox(
                  height: 16,
                  width: 16,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              ],
              if (_pending.isNotEmpty) ...[
                const SizedBox(height: 20),
                const Divider(height: 1),
                ConstrainedBox(
                  constraints: BoxConstraints(maxHeight: listMaxHeight),
                  child: Scrollbar(
                    controller: _pendingScrollController,
                    thumbVisibility: true,
                    child: ListView(
                      controller: _pendingScrollController,
                      shrinkWrap: true,
                      children: [
                        for (final recording in _pending)
                          _PendingRecordingTile(
                            recording: recording,
                            isPlaying: _playingPath == recording.localPath,
                            colorScheme: colorScheme,
                            l10n: l10n,
                            onPlay: recording.localPath != null
                                ? () => _playRecording(recording.localPath!)
                                : null,
                            onCancel: () => _cancelRecording(recording),
                          ),
                      ],
                    ),
                  ),
                ),
              ],
              if (_readyForReview.isNotEmpty) ...[
                const SizedBox(height: 20),
                const Divider(height: 1),
                Padding(
                  padding: const EdgeInsets.only(top: 12, bottom: 4),
                  child: Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      l10n.voiceReviewSectionTitle,
                      style: Theme.of(context).textTheme.labelLarge,
                    ),
                  ),
                ),
                ConstrainedBox(
                  constraints: BoxConstraints(maxHeight: listMaxHeight),
                  child: Scrollbar(
                    controller: _readyForReviewScrollController,
                    thumbVisibility: true,
                    child: ListView(
                      controller: _readyForReviewScrollController,
                      shrinkWrap: true,
                      children: [
                        for (final recording in _readyForReview)
                          _ReadyForReviewTile(
                            recording: recording,
                            colorScheme: colorScheme,
                            l10n: l10n,
                            onDismiss: () => _rejectRecording(recording),
                            onTap: () => _showRecordingDetail(recording),
                          ),
                      ],
                    ),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _PendingRecordingTile extends StatelessWidget {
  final VoiceRecording recording;
  final bool isPlaying;
  final ColorScheme colorScheme;
  final AppLocalizations l10n;
  final VoidCallback? onPlay;
  final VoidCallback onCancel;

  const _PendingRecordingTile({
    required this.recording,
    required this.isPlaying,
    required this.colorScheme,
    required this.l10n,
    required this.onPlay,
    required this.onCancel,
  });

  @override
  Widget build(BuildContext context) {
    final isProcessing =
        recording.status == voiceRecordingStatusProcessing;
    return ListTile(
      dense: true,
      leading: Icon(Icons.graphic_eq,
          color: isPlaying ? colorScheme.primary : null),
      title: Text(
        isProcessing
            ? l10n.voiceRecordingStatusProcessing
            : l10n.voiceRecordingStatusPending,
        style: isPlaying ? TextStyle(color: colorScheme.primary) : null,
      ),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          IconButton(
            icon: Icon(isPlaying ? Icons.stop : Icons.play_arrow,
                color: isPlaying ? colorScheme.primary : null),
            tooltip: isPlaying
                ? l10n.voiceRecordingStopTooltip
                : l10n.voiceRecordingPlayTooltip,
            onPressed: onPlay,
          ),
          if (!isProcessing)
            IconButton(
              icon: const Icon(Icons.close),
              tooltip: l10n.voiceRecordingCancelTooltip,
              onPressed: onCancel,
            ),
        ],
      ),
    );
  }
}

const _transcriptPreviewMaxLength = 60;

class _ReadyForReviewTile extends StatelessWidget {
  final VoiceRecording recording;
  final ColorScheme colorScheme;
  final AppLocalizations l10n;

  final VoidCallback onDismiss;
  final VoidCallback onTap;

  const _ReadyForReviewTile({
    required this.recording,
    required this.colorScheme,
    required this.l10n,
    required this.onDismiss,
    required this.onTap,
  });

  VoiceAction? get _errorAction {
    for (final action in recording.voiceActions) {
      if (action.status == 'error') return action;
    }
    return null;
  }

  String _errorLabel(String? code) {
    switch (code) {
      case 'HIVE_NOT_IDENTIFIED':
        return l10n.voiceReviewErrorHiveNotIdentified;
      case 'MULTIPLE_HIVES_MENTIONED':
        return l10n.voiceReviewErrorMultipleHives;
      case 'PROPOSAL_INCOMPLETE':
        return l10n.voiceReviewErrorProposalIncomplete;
      default:
        return l10n.voiceReviewErrorGeneric;
    }
  }

  @override
  Widget build(BuildContext context) {
    final errorAction = _errorAction;
    final noActionRecognized = errorAction == null && recording.voiceActions.isEmpty;
    final isError = errorAction != null || noActionRecognized;

    final transcript = recording.transcript?.trim();
    final title = (transcript != null && transcript.isNotEmpty)
        ? (transcript.length > _transcriptPreviewMaxLength
            ? '${transcript.substring(0, _transcriptPreviewMaxLength)}…'
            : transcript)
        : l10n.voiceRecordingStatusCompleted;

    final subtitle = errorAction != null
        ? _errorLabel(errorAction.errorMessage)
        : noActionRecognized
            ? l10n.voiceReviewNoActionRecognized
            : recording.createdAt != null
                ? DateFormat('d.MM HH:mm').format(recording.createdAt!)
                : null;

    return ListTile(
      dense: true,
      onTap: onTap,
      leading: Icon(
        isError ? Icons.error_outline : Icons.fact_check_outlined,
        color: isError ? colorScheme.error : colorScheme.primary,
      ),
      title: Text(title, maxLines: 1, overflow: TextOverflow.ellipsis),
      subtitle: subtitle != null
          ? Text(
              subtitle,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: isError ? TextStyle(color: colorScheme.error) : null,
            )
          : null,
      trailing: isError
          ? IconButton(
              icon: const Icon(Icons.close),
              tooltip: l10n.voiceReviewDismissTooltip,
              onPressed: onDismiss,
            )
          : null,
    );
  }
}

class _VoiceRecordingDetailDialog extends StatefulWidget {
  final int apiaryId;
  final VoiceRecording recording;
  final Future<bool> Function() onAccept;
  final Future<bool> Function() onReject;

  const _VoiceRecordingDetailDialog({
    required this.apiaryId,
    required this.recording,
    required this.onAccept,
    required this.onReject,
  });

  @override
  State<_VoiceRecordingDetailDialog> createState() =>
      _VoiceRecordingDetailDialogState();
}

class _VoiceRecordingDetailDialogState
    extends State<_VoiceRecordingDetailDialog> {
  bool _submitting = false;
  int? _loadingActionId;
  late VoiceRecording _recording = widget.recording;

  VoiceRecording get recording => _recording;

  Future<void> _handle(Future<bool> Function() action) async {
    setState(() => _submitting = true);
    final success = await action();
    if (!mounted) return;
    if (success) {
      Navigator.of(context).pop();
    } else {
      setState(() => _submitting = false);
    }
  }

  Future<void> _openResult(VoiceAction action) async {
    final resultType = action.resultType;
    final resultId = action.resultRecordId;
    final hiveId = action.hiveId;
    if (resultType == null || resultId == null || hiveId == null) return;

    final l10n = AppLocalizations.of(context)!;
    final api = context.read<ApiClient>();
    final apiaryId = widget.apiaryId;

    setState(() => _loadingActionId = action.id);
    try {
      final hive = await HiveRepository(api: api).getHive(apiaryId, hiveId);
      Widget screen;
      switch (resultType) {
        case 'inspection':
          final inspection = await InspectionRepository(api: api)
              .getInspection(
                apiaryId: apiaryId,
                hiveId: hiveId,
                inspectionId: resultId,
              );
          screen = InspectionFormScreen(
            apiaryId: apiaryId,
            hive: hive,
            inspection: inspection,
          );
          break;
        case 'treatment':
          final treatment = await TreatmentRepository(api: api).getTreatment(
            apiaryId: apiaryId,
            hiveId: hiveId,
            treatmentId: resultId,
          );
          screen = TreatmentFormScreen(
            apiaryId: apiaryId,
            hive: hive,
            treatment: treatment,
          );
          break;
        case 'harvest':
          final harvest = await HarvestRepository(api: api).getHarvest(
            apiaryId: apiaryId,
            hiveId: hiveId,
            harvestId: resultId,
          );
          screen = HarvestFormScreen(
            apiaryId: apiaryId,
            hive: hive,
            harvest: harvest,
          );
          break;
        case 'feeding':
          final feeding = await FeedingRepository(api: api).getFeeding(
            apiaryId: apiaryId,
            hiveId: hiveId,
            feedingId: resultId,
          );
          screen = FeedingFormScreen(
            apiaryId: apiaryId,
            hive: hive,
            feeding: feeding,
          );
          break;
        case 'hive_status':
          screen = EditHiveScreen(apiaryId: apiaryId, hive: hive);
          break;
        default:
          setState(() => _loadingActionId = null);
          return;
      }
      if (!mounted) return;
      final navigator = Navigator.of(context);
      setState(() => _loadingActionId = null);
      navigator.pop();
      navigator.push(MaterialPageRoute(builder: (_) => screen));
    } catch (_) {
      if (!mounted) return;
      setState(() => _loadingActionId = null);
      showBigSnackBar(context, l10n.generalError);
    }
  }

  static const _editableProposedTools = {
    'create_inspection',
    'create_treatment',
    'create_harvest',
    'create_feeding',
    'update_hive_status',
  };

  Future<void> _editProposedAction(VoiceAction action) async {
    final args = action.toolArguments;
    final hiveId = action.hiveId;
    if (args == null || hiveId == null) return;

    final l10n = AppLocalizations.of(context)!;
    final api = context.read<ApiClient>();
    final apiaryId = widget.apiaryId;

    setState(() => _loadingActionId = action.id);
    try {
      final hive = await HiveRepository(api: api).getHive(apiaryId, hiveId);
      if (!mounted) return;
      Widget screen;
      switch (action.toolName) {
        case 'create_inspection':
          screen = InspectionFormScreen(
            apiaryId: apiaryId,
            hive: hive,
            inspection: _inspectionFromArgs(args, hiveId),
            onSaveProposed: (edited) =>
                _persistProposedEdit(action, edited),
          );
          break;
        case 'create_treatment':
          screen = TreatmentFormScreen(
            apiaryId: apiaryId,
            hive: hive,
            treatment: _treatmentFromArgs(args, hiveId),
            onSaveProposed: (edited) =>
                _persistProposedEdit(action, edited),
          );
          break;
        case 'create_harvest':
          screen = HarvestFormScreen(
            apiaryId: apiaryId,
            hive: hive,
            harvest: _harvestFromArgs(args, hiveId),
            onSaveProposed: (edited) =>
                _persistProposedEdit(action, edited),
          );
          break;
        case 'create_feeding':
          screen = FeedingFormScreen(
            apiaryId: apiaryId,
            hive: hive,
            feeding: _feedingFromArgs(args, hiveId),
            onSaveProposed: (edited) =>
                _persistProposedEdit(action, edited),
          );
          break;
        case 'update_hive_status':
          screen = UpdateHiveStatusFormScreen(
            hive: hive,
            toolArguments: args,
            onSaveProposed: (edited) => _persistProposedEdit(action, edited),
          );
          break;
        default:
          setState(() => _loadingActionId = null);
          return;
      }
      await Navigator.of(context).push(
        MaterialPageRoute(builder: (_) => screen),
      );
      if (!mounted) return;
      setState(() => _loadingActionId = null);
    } catch (_) {
      if (!mounted) return;
      setState(() => _loadingActionId = null);
      showBigSnackBar(context, l10n.generalError);
    }
  }

  Future<void> _persistProposedEdit(
    VoiceAction action,
    Map<String, dynamic> toolArguments,
  ) async {
    final updated = await VoiceRepository(api: context.read<ApiClient>())
        .updateActionArguments(
      widget.apiaryId,
      recording.recordingId,
      action.id,
      toolArguments,
    );
    if (!mounted) return;
    setState(() {
      _recording = recording.copyWith(
        voiceActions: [
          for (final a in recording.voiceActions)
            if (a.id == action.id) updated else a,
        ],
      );
    });
  }

  String _toolLabel(AppLocalizations l10n, String? toolName) {
    switch (toolName) {
      case 'create_inspection':
        return l10n.voiceToolCreateInspection;
      case 'create_treatment':
        return l10n.voiceToolCreateTreatment;
      case 'create_harvest':
        return l10n.voiceToolCreateHarvest;
      case 'create_feeding':
        return l10n.voiceToolCreateFeeding;
      case 'update_hive_status':
        return l10n.voiceToolUpdateHiveStatus;
      default:
        return toolName ?? '';
    }
  }

  String _prettyKey(String key) {
    return key
        .split('_')
        .where((w) => w.isNotEmpty)
        .map((w) => w[0].toUpperCase() + w.substring(1))
        .join(' ');
  }

  String _prettyValue(AppLocalizations l10n, dynamic value) {
    if (value == null) return '';
    if (value is bool) return value ? l10n.generalYes : l10n.generalNo;
    if (value is List) return value.join(', ');
    return value.toString();
  }

  Inspection _inspectionFromArgs(Map<String, dynamic> args, int hiveId) {
    return Inspection(
      id: 0,
      hiveId: hiveId,
      inspectedAt: DateTime.now(),
      queenSeen: args['queen_status'] as String? ?? '',
      broodPattern: args['brood_pattern'] as String? ?? '',
      aggressiveness: args['aggressiveness'] as String? ?? '',
      colonyStrength: args['colony_strength'] as String? ?? '',
      framesBrood: (args['frames_brood'] as num?)?.toInt(),
      framesFeed: (args['frames_feed'] as num?)?.toInt(),
      framesPollen: (args['frames_pollen'] as num?)?.toInt(),
      framesAddedDrawn: (args['frames_added_drawn'] as num?)?.toInt(),
      framesAddedFoundation: (args['frames_added_foundation'] as num?)?.toInt(),
      framesAddedBrood: (args['frames_added_brood'] as num?)?.toInt(),
      framesAddedFeed: (args['frames_added_feed'] as num?)?.toInt(),
      queenCellsCount: (args['queen_cells_count'] as num?)?.toInt(),
      queenAdded: args['queen_added'] as bool? ?? false,
      boxAdded: args['box_added'] as bool? ?? false,
      notes: args['notes'] as String? ?? '',
    );
  }

  Treatment _treatmentFromArgs(Map<String, dynamic> args, int hiveId) {
    return Treatment(
      id: 0,
      hiveId: hiveId,
      treatedAt: DateTime.now(),
      medicineName: args['medicine_name'] as String? ?? '',
      dose: args['dose'] as String? ?? '1',
      notes: args['notes'] as String? ?? '',
    );
  }

  Harvest _harvestFromArgs(Map<String, dynamic> args, int hiveId) {
    return Harvest(
      id: 0,
      hiveId: hiveId,
      harvestedAt: DateTime.now(),
      frames: (args['frames'] as num?)?.toInt() ?? 0,
      halfFrames: (args['half_frames'] as num?)?.toInt() ?? 0,
      kilograms: (args['kilograms'] as num?)?.toDouble() ?? 0,
      notes: args['notes'] as String? ?? '',
    );
  }

  Feeding _feedingFromArgs(Map<String, dynamic> args, int hiveId) {
    return Feeding(
      id: 0,
      hiveId: hiveId,
      fedAt: DateTime.now(),
      feedType: args['feed_type'] as String? ?? '',
      amount: args['amount'] as String? ?? '',
      notes: args['notes'] as String? ?? '',
    );
  }

  String _errorLabel(AppLocalizations l10n, String? code) {
    switch (code) {
      case 'HIVE_NOT_IDENTIFIED':
        return l10n.voiceReviewErrorHiveNotIdentified;
      case 'MULTIPLE_HIVES_MENTIONED':
        return l10n.voiceReviewErrorMultipleHives;
      case 'PROPOSAL_INCOMPLETE':
        return l10n.voiceReviewErrorProposalIncomplete;
      default:
        return l10n.voiceReviewErrorGeneric;
    }
  }

  Widget _diseasesField(
    BuildContext context,
    AppLocalizations l10n,
    List<String> diseases,
  ) {
    return Padding(
      padding: const EdgeInsets.only(top: 6),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            l10n.inspectionDiseases,
            style: Theme.of(context).textTheme.labelSmall?.copyWith(
                  color: Theme.of(context).colorScheme.primary,
                  fontWeight: FontWeight.w600,
                  letterSpacing: 0.4,
                ),
          ),
          const SizedBox(height: 2),
          Text(
            diseases.map((d) => hiveDiseaseLabel(l10n, d)).join(' · '),
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Theme.of(context).colorScheme.onSurfaceVariant,
                ),
          ),
        ],
      ),
    );
  }

  List<Widget> _fieldRows(List<(String, String?)> fields) {
    return [
      for (final (label, value) in fields)
        if (value != null && value.isNotEmpty)
          Padding(
            padding: const EdgeInsets.only(top: 2),
            child: Text('$label: $value'),
          ),
    ];
  }

  IconData _actionIcon(String? toolName) {
    switch (toolName) {
      case 'create_inspection':
        return Icons.search;
      case 'create_treatment':
        return Icons.medical_services_outlined;
      case 'create_harvest':
        return Icons.water_drop_outlined;
      case 'create_feeding':
        return Icons.grain;
      case 'update_hive_status':
        return Icons.flag_outlined;
      default:
        return Icons.check_circle_outline;
    }
  }

  Widget _buildActionCard(
    BuildContext context,
    AppLocalizations l10n,
    ColorScheme colorScheme,
    VoiceAction action,
  ) {
    final isEditable = (action.status == 'applied' &&
            action.resultRecordId != null) ||
        (action.status == 'proposed' &&
            _editableProposedTools.contains(action.toolName));
    final isLoading = _loadingActionId == action.id;

    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: !isEditable || _loadingActionId != null
            ? null
            : action.status == 'applied'
                ? () => _openResult(action)
                : () => _editProposedAction(action),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(
                    _actionIcon(action.toolName),
                    size: 20,
                    color: colorScheme.primary,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      _toolLabel(l10n, action.toolName),
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                  ),
                  if (isLoading)
                    const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  else if (isEditable)
                    Icon(
                      Icons.chevron_right,
                      color: colorScheme.onSurfaceVariant,
                    ),
                ],
              ),
              const SizedBox(height: 12),
              _buildActionBody(context, l10n, action),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildActionBody(
    BuildContext context,
    AppLocalizations l10n,
    VoiceAction action,
  ) {
    final args = action.toolArguments ?? {};
    final diseases = (args['diseases'] as List?)?.cast<String>() ?? const [];

    switch (action.toolName) {
      case 'create_inspection':
        return InspectionSummary(
          inspection: _inspectionFromArgs(args, action.hiveId ?? 0),
        );
      case 'create_treatment':
        return TreatmentSummary(
          treatment: _treatmentFromArgs(args, action.hiveId ?? 0),
          l10n: l10n,
          showDate: false,
        );
      case 'create_harvest':
        return HarvestSummary(
          harvest: _harvestFromArgs(args, action.hiveId ?? 0),
          l10n: l10n,
          showDate: false,
        );
      case 'create_feeding':
        return FeedingSummary(
          feeding: _feedingFromArgs(args, action.hiveId ?? 0),
          l10n: l10n,
          showDate: false,
        );
      case 'update_hive_status':
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            ..._fieldRows([
              if (args['ready_for_harvest'] != null)
                (
                  l10n.hiveReadyForHarvest,
                  args['ready_for_harvest'] as bool
                      ? l10n.generalYes
                      : l10n.generalNo,
                ),
              if (args['queen_needs_replacement'] != null)
                (
                  l10n.hiveQueenNeedsReplacement,
                  args['queen_needs_replacement'] as bool
                      ? l10n.generalYes
                      : l10n.generalNo,
                ),
              if (args['needs_food'] != null)
                (
                  l10n.hiveNeedsFood,
                  args['needs_food'] as bool
                      ? l10n.generalYes
                      : l10n.generalNo,
                ),
              if (args['box_needs_adding'] != null)
                (
                  l10n.hiveBoxNeedsAdding,
                  args['box_needs_adding'] as bool
                      ? l10n.generalYes
                      : l10n.generalNo,
                ),
            ]),
            if (diseases.isNotEmpty) _diseasesField(context, l10n, diseases),
          ],
        );
      default:
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            for (final entry in args.entries)
              Padding(
                padding: const EdgeInsets.only(top: 2),
                child: Text(
                  '${_prettyKey(entry.key)}: ${_prettyValue(l10n, entry.value)}',
                ),
              ),
          ],
        );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final transcript = recording.transcript?.trim();
    final hasProposedAction =
        recording.voiceActions.any((a) => a.status == 'proposed');

    return AlertDialog(
      title: Text(l10n.voiceReviewDetailTitle),
      content: SizedBox(
        width: AppLayout.dialogWidth(context),
        child: _buildDialogContent(context, l10n, colorScheme, transcript),
      ),
      actions: [
        TextButton(
          onPressed: _submitting ? null : () => Navigator.of(context).pop(),
          child: Text(l10n.generalClose),
        ),
        TextButton(
          onPressed: _submitting ? null : () => _handle(widget.onReject),
          style: TextButton.styleFrom(foregroundColor: colorScheme.error),
          child: Text(l10n.voiceReviewRejectAction),
        ),
        if (hasProposedAction)
          FilledButton(
            onPressed: _submitting ? null : () => _handle(widget.onAccept),
            child: Text(l10n.voiceReviewAcceptAction),
          ),
      ],
    );
  }

  Widget _buildDialogContent(
    BuildContext context,
    AppLocalizations l10n,
    ColorScheme colorScheme,
    String? transcript,
  ) {
    return SizedBox(
      width: AppLayout.dialogWidth(context),
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            if (transcript != null && transcript.isNotEmpty)
              SelectableText(transcript)
            else
              Text(l10n.voiceRecordingStatusCompleted),
            const SizedBox(height: 16),
            Text(l10n.voiceReviewDetailActionsLabel,
                style: Theme.of(context).textTheme.labelLarge),
            const SizedBox(height: 8),
            if (recording.voiceActions.isEmpty)
              Text(
                l10n.voiceReviewNoActionRecognized,
                style: TextStyle(color: colorScheme.error),
              )
            else
              for (final action in recording.voiceActions)
                Padding(
                  padding: const EdgeInsets.only(bottom: 12),
                  child: action.status == 'error'
                      ? Text(
                          _errorLabel(l10n, action.errorMessage),
                          style: TextStyle(color: colorScheme.error),
                        )
                      : _buildActionCard(context, l10n, colorScheme, action),
                ),
          ],
        ),
      ),
    );
  }
}

class _HiveStatusIcons extends StatelessWidget {
  final Hive hive;

  const _HiveStatusIcons({required this.hive});

  @override
  Widget build(BuildContext context) {
    final color = Theme.of(context).colorScheme.onSurfaceVariant;
    final icons = [
      if (hive.queenNeedsReplacement)
        (Icons.published_with_changes_outlined, color),
      if (hive.readyForHarvest) (Icons.water_drop_outlined, Colors.amber.shade700),
      if (hive.needsFood) (Icons.restaurant_outlined, Colors.orange.shade700),
      if (hive.boxNeedsAdding) (Icons.inventory_2_outlined, Colors.brown.shade400),
      if (hive.diseases.isNotEmpty) (Icons.coronavirus_outlined, Colors.red.shade400),
    ];
    if (icons.isEmpty) return const SizedBox.shrink();
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: icons
          .map((e) => Icon(e.$1, size: 26, color: e.$2))
          .toList(),
    );
  }
}

class _HiveCell extends StatelessWidget {
  final Hive hive;
  final bool dimmed;
  final VoidCallback onTap;

  const _HiveCell({
    required this.hive,
    required this.dimmed,
    required this.onTap,
  });

  Widget _buildContent(BuildContext context, {double opacity = 1.0}) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final primary = colorScheme.primary;
    final locale = Localizations.localeOf(context).toString();

    return Opacity(
      opacity: opacity,
      child: Card(
        margin: EdgeInsets.zero,
        color: primary.withAlpha(40),
        clipBehavior: Clip.antiAlias,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
          side: const BorderSide(color: Colors.black),
        ),
        child: Padding(
          padding: const EdgeInsets.all(6),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.hive,
                size: 24,
                color: hive.active ? primary : colorScheme.outline,
              ),
              const SizedBox(height: 2),
              Flexible(
                child: Text(
                  hive.name,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                        fontSize: 12,
                      ),
                  textAlign: TextAlign.center,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (!hive.active)
                Text(
                  l10n.hiveInactive,
                  style: Theme.of(context).textTheme.labelSmall?.copyWith(
                        color: colorScheme.outline,
                        fontSize: 9,
                      ),
                ),
              if (hive.lastInspectedAt != null)
                Text(
                  DateFormat('d MMM', locale).format(hive.lastInspectedAt!),
                  style: Theme.of(context).textTheme.labelSmall?.copyWith(
                        color: colorScheme.onSurfaceVariant,
                        fontSize: 9,
                      ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final effectiveOpacity = dimmed ? 0.3 : 1.0;
    return LongPressDraggable<Hive>(
      data: hive,
      delay: const Duration(milliseconds: 300),
      feedback: SizedBox(
        width: 80,
        height: 80,
        child: Material(
          color: Colors.transparent,
          elevation: 8,
          borderRadius: BorderRadius.circular(12),
          child: _buildContent(context),
        ),
      ),
      childWhenDragging: _buildContent(context, opacity: 0.35),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: _buildContent(context, opacity: effectiveOpacity),
      ),
    );
  }
}

class _EmptyCell extends StatelessWidget {
  final VoidCallback? onTap;
  final void Function(Hive)? onDrop;

  const _EmptyCell({this.onTap, this.onDrop});

  @override
  Widget build(BuildContext context) {
    return DragTarget<Hive>(
      onWillAcceptWithDetails: (_) => true,
      onAcceptWithDetails: (details) => onDrop?.call(details.data),
      builder: (context, candidateData, _) {
        final hovering = candidateData.isNotEmpty;
        return InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(12),
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 150),
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: hovering
                    ? Theme.of(context).colorScheme.primary
                    : Theme.of(context).colorScheme.outlineVariant,
                width: hovering ? 2 : 1,
              ),
              color: hovering
                  ? Theme.of(context).colorScheme.primary.withAlpha(20)
                  : null,
            ),
          ),
        );
      },
    );
  }
}
