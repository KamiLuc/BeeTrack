import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

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
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.hive.name);
    _type = widget.hive.type;
    _active = widget.hive.active;
    _queenless = widget.hive.queenless;
    _readyForHarvest = widget.hive.readyForHarvest;
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  Future<void> _submit(BuildContext context) async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() => _loading = true);
    try {
      await HiveRepository(api: context.read()).updateHive(
        apiaryId: widget.apiaryId,
        hiveId: widget.hive.id,
        name: _nameController.text.trim(),
        type: _type,
        active: _active,
        queenless: _queenless,
        readyForHarvest: _readyForHarvest,
      );
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
