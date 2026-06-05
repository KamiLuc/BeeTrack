import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../l10n/app_localizations.dart';
import '../data/hive_model.dart';
import '../data/hive_repository.dart';
import 'hive_form_widgets.dart';

class EditHiveScreen extends StatefulWidget {
  final int apiaryId;
  final Hive hive;

  const EditHiveScreen({
    super.key,
    required this.apiaryId,
    required this.hive,
  });

  @override
  State<EditHiveScreen> createState() => _EditHiveScreenState();
}

class _EditHiveScreenState extends State<EditHiveScreen> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _nameController;
  late String _type;
  late bool _active;
  late bool _queenless;
  late bool _readyForHarvest;
  late Set<String> _hiveDiseases;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.hive.name);
    _type = widget.hive.type;
    _active = widget.hive.active;
    _queenless = widget.hive.queenless;
    _readyForHarvest = widget.hive.readyForHarvest;
    _hiveDiseases = widget.hive.diseases.map((d) => d.disease).toSet();
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  Future<void> _submit(BuildContext context) async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() => _loading = true);
    final repo = HiveRepository(api: context.read<ApiClient>());
    try {
      await repo.updateHive(
        apiaryId: widget.apiaryId,
        hiveId: widget.hive.id,
        name: _nameController.text.trim(),
        type: _type,
        active: _active,
        queenless: _queenless,
        readyForHarvest: _readyForHarvest,
      );

      final existing = widget.hive.diseases.map((d) => d.disease).toSet();
      final toAdd = _hiveDiseases.difference(existing);
      final toRemove = existing.difference(_hiveDiseases);

      final keptDiseases = widget.hive.diseases
          .where((d) => !toRemove.contains(d.disease))
          .toList();
      final addedDiseases = <HiveDisease>[];
      for (final disease in toAdd) {
        final d = await repo.addDisease(
          apiaryId: widget.apiaryId,
          hiveId: widget.hive.id,
          disease: disease,
        );
        addedDiseases.add(d);
      }
      for (final disease in toRemove) {
        final d = widget.hive.diseases.firstWhere((d) => d.disease == disease);
        await repo.removeDisease(
          apiaryId: widget.apiaryId,
          hiveId: widget.hive.id,
          diseaseId: d.id,
        );
      }

      if (context.mounted) {
        Navigator.of(context).pop(Hive(
          id: widget.hive.id,
          apiaryId: widget.hive.apiaryId,
          name: _nameController.text.trim(),
          type: _type,
          active: _active,
          queenless: _queenless,
          readyForHarvest: _readyForHarvest,
          gridRow: widget.hive.gridRow,
          gridCol: widget.hive.gridCol,
          diseases: [...keptDiseases, ...addedDiseases],
          lastInspectedAt: widget.hive.lastInspectedAt,
        ));
      }
    } catch (_) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.generalError)),
        );
      }
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.hiveEdit)),
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
                    HiveNameField(controller: _nameController),
                    const SizedBox(height: 16),
                    HiveTypeDropdown(
                      value: _type,
                      onChanged: (v) => setState(() => _type = v ?? _type),
                    ),
                    const SizedBox(height: 16),
                    HiveActiveToggle(
                      value: _active,
                      onChanged: (v) => setState(() => _active = v),
                    ),
                    const SizedBox(height: 16),
                    HiveQueenlessToggle(
                      value: _queenless,
                      onChanged: (v) => setState(() => _queenless = v),
                    ),
                    const SizedBox(height: 16),
                    HiveReadyForHarvestToggle(
                      value: _readyForHarvest,
                      onChanged: (v) => setState(() => _readyForHarvest = v),
                    ),
                    const SizedBox(height: 16),
                    HiveDiseasesSection(
                      label: AppLocalizations.of(context)!.inspectionDiseases,
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
