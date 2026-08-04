import 'package:flutter/material.dart';

import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../data/hive_model.dart';
import 'hive_form_widgets.dart';

class UpdateHiveStatusFormScreen extends StatefulWidget {
  final Hive hive;
  final Map<String, dynamic> toolArguments;
  final Future<void> Function(Map<String, dynamic> toolArguments)
  onSaveProposed;

  const UpdateHiveStatusFormScreen({
    super.key,
    required this.hive,
    required this.toolArguments,
    required this.onSaveProposed,
  });

  @override
  State<UpdateHiveStatusFormScreen> createState() =>
      _UpdateHiveStatusFormScreenState();
}

class _UpdateHiveStatusFormScreenState
    extends State<UpdateHiveStatusFormScreen> {
  late bool _readyForHarvest;
  late bool _queenNeedsReplacement;
  late bool _needsFood;
  late bool _boxNeedsAdding;
  late Set<String> _diseases;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    final args = widget.toolArguments;
    final hive = widget.hive;
    _readyForHarvest =
        args['ready_for_harvest'] as bool? ?? hive.readyForHarvest;
    _queenNeedsReplacement =
        args['queen_needs_replacement'] as bool? ?? hive.queenNeedsReplacement;
    _needsFood = args['needs_food'] as bool? ?? hive.needsFood;
    _boxNeedsAdding = args['box_needs_adding'] as bool? ?? hive.boxNeedsAdding;
    _diseases = (args['diseases'] as List?)?.cast<String>().toSet() ??
        hive.diseases.map((d) => d.disease).toSet();
  }

  Future<void> _submit(BuildContext ctx) async {
    setState(() => _loading = true);
    try {
      await widget.onSaveProposed({
        'ready_for_harvest': _readyForHarvest,
        'queen_needs_replacement': _queenNeedsReplacement,
        'needs_food': _needsFood,
        'box_needs_adding': _boxNeedsAdding,
        'diseases': _diseases.toList(),
      });
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

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.voiceToolUpdateHiveStatus),
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
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Text(
                          l10n.voiceUpdateHiveStatusCurrentStateHint,
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: Theme.of(context)
                                    .colorScheme
                                    .onSurfaceVariant,
                              ),
                        ),
                        const SizedBox(height: 20),
                        HiveSectionTitle(l10n.inspectionSectionTodo),
                        const SizedBox(height: 12),
                        HiveReadyForHarvestToggle(
                          value: _readyForHarvest,
                          onChanged: (v) =>
                              setState(() => _readyForHarvest = v),
                        ),
                        const SizedBox(height: 16),
                        HiveNeedsFoodToggle(
                          value: _needsFood,
                          onChanged: (v) => setState(() => _needsFood = v),
                        ),
                        const SizedBox(height: 16),
                        HiveQueenNeedsReplacementToggle(
                          value: _queenNeedsReplacement,
                          onChanged: (v) =>
                              setState(() => _queenNeedsReplacement = v),
                        ),
                        const SizedBox(height: 16),
                        HiveBoxNeedsAddingToggle(
                          value: _boxNeedsAdding,
                          onChanged: (v) =>
                              setState(() => _boxNeedsAdding = v),
                        ),
                        const SizedBox(height: 20),
                        HiveDiseasesSection(
                          label: l10n.inspectionDiseases,
                          selected: _diseases,
                          onToggle: (disease, selected) {
                            setState(() {
                              if (selected) {
                                _diseases = {..._diseases, disease};
                              } else {
                                _diseases = _diseases
                                    .where((d) => d != disease)
                                    .toSet();
                              }
                            });
                          },
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          ),
          SafeArea(
            top: false,
            child: Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Center(
                child: SizedBox(
                  width: AppLayout.bannerWidth(context),
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
                        _loading
                            ? const Padding(
                                padding: EdgeInsets.all(12),
                                child: SizedBox(
                                  width: 24,
                                  height: 24,
                                  child: CircularProgressIndicator(
                                      strokeWidth: 2),
                                ),
                              )
                            : IconButton(
                                icon: const Icon(Icons.check),
                                iconSize: 28,
                                tooltip: l10n.generalSave,
                                onPressed: () => _submit(context),
                              ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
