import 'package:flutter/material.dart';
import '../../../core/widgets/profile_icon_button.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/app_layout.dart';
import '../../../l10n/app_localizations.dart';
import '../../hive/data/hive_model.dart';
import '../../hive/data/hive_repository.dart';
import '../../hive/view/hive_form_widgets.dart';
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

  // Fields pre-filled from the previous inspection are shown in grey until
  // the user modifies them.
  final Set<String> _defaultFields = {};

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

    _framesBroodController = _initFrameCtrl('framesBrood', insp?.framesBrood, prev?.framesBrood);
    _framesHoneyController = _initFrameCtrl('framesHoney', insp?.framesHoney, prev?.framesHoney);
    _framesPollenController = _initFrameCtrl('framesPollen', insp?.framesPollen, prev?.framesPollen);

    _framesAddedDrawnController = TextEditingController(
      text: insp?.framesAddedDrawn?.toString() ?? '',
    );
    _framesAddedFoundationController = TextEditingController(
      text: insp?.framesAddedFoundation?.toString() ?? '',
    );
    _framesAddedHoneyController = TextEditingController(
      text: insp?.framesAddedHoney?.toString() ?? '',
    );
    _queenCellsCountController = TextEditingController(
      text: insp?.queenCellsCount?.toString() ?? '',
    );
    _notesController = TextEditingController(text: insp?.notes ?? '');

    _hiveActive = widget.hive.active;
    _hiveQueenless = widget.hive.queenless;
    _hiveReadyForHarvest = widget.hive.readyForHarvest;
    _hiveDiseases = widget.hive.diseases.map((d) => d.disease).toSet();
  }

  TextEditingController _initFrameCtrl(
    String fieldKey,
    int? editValue,
    int? prevValue,
  ) {
    final text = editValue?.toString() ?? prevValue?.toString() ?? '';
    if (prevValue != null && editValue == null) _defaultFields.add(fieldKey);
    return TextEditingController(text: text);
  }

  void _clearDefault(String fieldKey) {
    if (_defaultFields.contains(fieldKey)) {
      setState(() => _defaultFields.remove(fieldKey));
    }
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
    final inspRepo = InspectionRepository(api: ctx.read());
    final hiveRepo = HiveRepository(api: ctx.read());
    try {
      if (widget.isEditing) {
        await inspRepo.updateInspection(
          apiaryId: widget.apiaryId,
          hiveId: widget.hive.id,
          inspectionId: widget.inspection!.id,
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
        await inspRepo.createInspection(
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
      }
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
    final textTheme = Theme.of(context).textTheme;
    return Scaffold(
      appBar: AppBar(
        title: Text(
          widget.isEditing ? l10n.inspectionEdit : l10n.inspectionAdd,
        ),
        actions: const [ProfileIconButton()],
      ),
      body: SafeArea(
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
                      onChanged: (v) => setState(() => _broodPattern = v ?? ''),
                    ),
                    const SizedBox(height: 16),
                    _EnumDropdown(
                      label: l10n.inspectionAggressiveness,
                      value: _aggressiveness.isEmpty ? null : _aggressiveness,
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
                    _NumericField(
                      controller: _framesBroodController,
                      label: l10n.inspectionFramesBrood,
                      isDefault: _defaultFields.contains('framesBrood'),
                      onModified: () => _clearDefault('framesBrood'),
                    ),
                    const SizedBox(height: 16),
                    _NumericField(
                      controller: _framesHoneyController,
                      label: l10n.inspectionFramesHoney,
                      isDefault: _defaultFields.contains('framesHoney'),
                      onModified: () => _clearDefault('framesHoney'),
                    ),
                    const SizedBox(height: 16),
                    _NumericField(
                      controller: _framesPollenController,
                      label: l10n.inspectionFramesPollen,
                      isDefault: _defaultFields.contains('framesPollen'),
                      onModified: () => _clearDefault('framesPollen'),
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
                      onChanged: (v) => setState(() => _hiveQueenless = v),
                    ),
                    const SizedBox(height: 12),
                    _BoolRow(
                      label: l10n.hiveReadyForHarvest,
                      value: _hiveReadyForHarvest,
                      onChanged: (v) => setState(() => _hiveReadyForHarvest = v),
                    ),
                    const SizedBox(height: 12),
                    _BoolRow(
                      label: l10n.inspectionQueenAdded,
                      value: _queenAdded,
                      onChanged: (v) => setState(() => _queenAdded = v),
                    ),
                    const SizedBox(height: 12),
                    HiveDiseasesSection(
                      label: l10n.inspectionDiseases,
                      selected: _hiveDiseases,
                      onToggle: (disease, selected) {
                        setState(() {
                          if (selected) {
                            _hiveDiseases = {..._hiveDiseases, disease};
                          } else {
                            _hiveDiseases = _hiveDiseases
                                .where((d) => d != disease)
                                .toSet();
                          }
                        });
                      },
                    ),
                    const SizedBox(height: 24),
                    Center(
                      child: SizedBox(
                        width: 200,
                        child: ElevatedButton(
                          onPressed: _loading ? null : () => _submit(context),
                          child: _loading
                              ? const SizedBox(
                                  height: 20,
                                  width: 20,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                    color: Colors.white,
                                  ),
                                )
                              : Text(l10n.generalSave),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
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
  final bool isDefault;
  final VoidCallback? onModified;

  const _NumericField({
    required this.controller,
    required this.label,
    this.isDefault = false,
    this.onModified,
  });

  @override
  Widget build(BuildContext context) {
    final color = isDefault
        ? Theme.of(context).colorScheme.onSurfaceVariant
        : null;
    return TextFormField(
      controller: controller,
      style: color != null ? TextStyle(color: color) : null,
      decoration: InputDecoration(
        labelText: label,
        border: const OutlineInputBorder(),
      ),
      keyboardType: TextInputType.number,
      inputFormatters: [FilteringTextInputFormatter.digitsOnly],
      onChanged: isDefault ? (_) => onModified?.call() : null,
    );
  }
}

