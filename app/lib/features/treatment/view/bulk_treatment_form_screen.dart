import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../data/treatment_repository.dart';

class BulkTreatmentFormScreen extends StatefulWidget {
  final int apiaryId;

  const BulkTreatmentFormScreen({super.key, required this.apiaryId});

  @override
  State<BulkTreatmentFormScreen> createState() => _BulkTreatmentFormScreenState();
}

class _BulkTreatmentFormScreenState extends State<BulkTreatmentFormScreen> {
  final _formKey = GlobalKey<FormState>();

  late DateTime _treatedAt;
  late final TextEditingController _medicineController;
  late final TextEditingController _doseController;
  late final TextEditingController _notesController;

  List<String> _medicineOptions = [];
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _treatedAt = DateTime.now();
    _medicineController = TextEditingController();
    _doseController = TextEditingController(text: '1');
    _notesController = TextEditingController();
    _loadMedicines();
  }

  Future<void> _loadMedicines() async {
    try {
      final options = await TreatmentRepository(api: context.read<ApiClient>()).listMedicines();
      if (mounted) setState(() => _medicineOptions = options);
    } catch (_) {}
  }

  @override
  void dispose() {
    _medicineController.dispose();
    _doseController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _treatedAt,
      firstDate: DateTime(2000),
      lastDate: DateTime.now(),
    );
    if (picked != null && mounted) {
      setState(
        () => _treatedAt = DateTime(
          picked.year,
          picked.month,
          picked.day,
          _treatedAt.hour,
          _treatedAt.minute,
        ),
      );
    }
  }

  Future<void> _submit(BuildContext ctx) async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() => _loading = true);
    final repo = TreatmentRepository(api: ctx.read<ApiClient>());
    try {
      final count = await repo.bulkTreatment(
        apiaryId: widget.apiaryId,
        treatedAt: _treatedAt,
        medicineName: _medicineController.text.trim(),
        dose: _doseController.text.trim(),
        notes: _notesController.text.trim(),
      );
      if (!ctx.mounted) return;
      Navigator.of(ctx).pop(count);
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
    final dateStr = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).format(_treatedAt);

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.treatmentTreatAllHives),
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
                          InkWell(
                            onTap: _pickDate,
                            child: InputDecorator(
                              decoration: InputDecoration(
                                labelText: l10n.treatmentDate,
                                suffixIcon: const Icon(Icons.calendar_today, size: 20),
                              ),
                              child: Text(dateStr),
                            ),
                          ),
                          const SizedBox(height: 16),
                          Autocomplete<String>(
                            initialValue: TextEditingValue(text: _medicineController.text),
                            optionsBuilder: (textEditingValue) {
                              if (textEditingValue.text.isEmpty) return _medicineOptions;
                              final query = textEditingValue.text.toLowerCase();
                              return _medicineOptions.where((m) => m.toLowerCase().contains(query));
                            },
                            onSelected: (selection) {
                              _medicineController.text = selection;
                            },
                            fieldViewBuilder: (context, controller, focusNode, onFieldSubmitted) {
                              if (controller.text.isEmpty && _medicineController.text.isNotEmpty) {
                                controller.text = _medicineController.text;
                              }
                              return TextFormField(
                                controller: controller,
                                focusNode: focusNode,
                                decoration: InputDecoration(labelText: l10n.treatmentMedicine),
                                onChanged: (v) => _medicineController.text = v,
                                validator: (v) => (v == null || v.trim().isEmpty)
                                    ? l10n.treatmentMedicineRequired
                                    : null,
                              );
                            },
                          ),
                          const SizedBox(height: 16),
                          TextFormField(
                            controller: _doseController,
                            decoration: InputDecoration(labelText: l10n.treatmentDose),
                            keyboardType: const TextInputType.numberWithOptions(decimal: true),
                            inputFormatters: [
                              FilteringTextInputFormatter.allow(RegExp(r'[0-9.,]')),
                            ],
                            validator: (v) => (v == null || v.trim().isEmpty)
                                ? l10n.treatmentDoseRequired
                                : null,
                          ),
                          const SizedBox(height: 16),
                          TextFormField(
                            controller: _notesController,
                            decoration: InputDecoration(labelText: l10n.treatmentNote),
                            maxLines: 3,
                            minLines: 1,
                          ),
                        ],
                      ),
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
                                  child: CircularProgressIndicator(strokeWidth: 2),
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
