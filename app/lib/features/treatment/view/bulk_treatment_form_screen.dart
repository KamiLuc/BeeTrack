import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../data/treatment_repository.dart';
import 'treatment_form_fields.dart';

class BulkTreatmentFormScreen extends StatefulWidget {
  final int apiaryId;
  final List<int>? hiveIds;

  const BulkTreatmentFormScreen({super.key, required this.apiaryId, this.hiveIds});

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
  List<String> _doseOptions = [];
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _treatedAt = DateTime.now();
    _medicineController = TextEditingController();
    _doseController = TextEditingController(text: '1');
    _notesController = TextEditingController();
    _loadSuggestions();
  }

  Future<void> _loadSuggestions() async {
    final repo = TreatmentRepository(api: context.read<ApiClient>());
    try {
      final options = await repo.listMedicines();
      if (mounted) setState(() => _medicineOptions = options);
    } catch (_) {}
    try {
      final doses = await repo.listDoses();
      if (mounted) setState(() => _doseOptions = doses);
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
    if (picked == null || !mounted) return;

    final pickedTime = await showTimePicker(
      context: context,
      initialTime: TimeOfDay.fromDateTime(_treatedAt),
    );
    if (!mounted) return;
    setState(
      () => _treatedAt = DateTime(
        picked.year,
        picked.month,
        picked.day,
        pickedTime?.hour ?? _treatedAt.hour,
        pickedTime?.minute ?? _treatedAt.minute,
      ),
    );
  }

  Future<void> _submit(BuildContext ctx) async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() => _loading = true);
    final repo = TreatmentRepository(api: ctx.read<ApiClient>());
    try {
      final count = await repo.bulkTreatment(
        apiaryId: widget.apiaryId,
        hiveIds: widget.hiveIds,
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
                    child: TreatmentFormFields(
                      formKey: _formKey,
                      treatedAt: _treatedAt,
                      medicineController: _medicineController,
                      doseController: _doseController,
                      notesController: _notesController,
                      medicineOptions: _medicineOptions,
                      doseOptions: _doseOptions,
                      onDateTap: _pickDate,
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
