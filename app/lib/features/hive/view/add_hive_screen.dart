import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/theme/app_layout.dart';
import '../../../l10n/app_localizations.dart';
import '../data/hive_repository.dart';
import 'hive_form_widgets.dart';

String _lastHiveType = 'langstroth';

class AddHiveScreen extends StatefulWidget {
  final int apiaryId;
  final int gridRow;
  final int gridCol;
  final String defaultName;

  const AddHiveScreen({
    super.key,
    required this.apiaryId,
    required this.gridRow,
    required this.gridCol,
    required this.defaultName,
  });

  @override
  State<AddHiveScreen> createState() => _AddHiveScreenState();
}

class _AddHiveScreenState extends State<AddHiveScreen> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _nameController;
  String _type = _lastHiveType;
  bool _active = true;
  bool _queenless = false;
  bool _readyForHarvest = false;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.defaultName);
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
      await HiveRepository(api: context.read()).createHive(
        apiaryId: widget.apiaryId,
        name: _nameController.text.trim(),
        type: _type,
        active: _active,
        queenless: _queenless,
        readyForHarvest: _readyForHarvest,
        gridRow: widget.gridRow,
        gridCol: widget.gridCol,
      );
      if (context.mounted) Navigator.of(context).pop();
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
      appBar: AppBar(title: Text(l10n.hiveAdd)),
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
                      onChanged: (v) => setState(() {
                        _type = v ?? _type;
                        _lastHiveType = _type;
                      }),
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
