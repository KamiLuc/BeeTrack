import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/validation/size_tiers.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../data/honey_batch_repository.dart';
import '../data/processing_method.dart';

class CreateHoneyBatchScreen extends StatefulWidget {
  const CreateHoneyBatchScreen({super.key});

  @override
  State<CreateHoneyBatchScreen> createState() => _CreateHoneyBatchScreenState();
}

class _CreateHoneyBatchScreenState extends State<CreateHoneyBatchScreen> {
  final _formKey = GlobalKey<FormState>();
  final _amountController = TextEditingController();
  final _honeyTypeController = TextEditingController();

  DateTime _gatheringDate = DateTime.now();
  ProcessingMethod _processingMethod = ProcessingMethod.raw;
  PlatformFile? _pickedFile;
  bool _requestCertification = false;
  bool _saving = false;
  String? _pdfError;

  @override
  void dispose() {
    _amountController.dispose();
    _honeyTypeController.dispose();
    super.dispose();
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _gatheringDate,
      firstDate: DateTime(2000),
      lastDate: DateTime.now(),
    );
    if (picked == null || !mounted) return;
    setState(() => _gatheringDate = picked);
  }

  Future<void> _pickPdf() async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['pdf'],
      withData: true,
    );
    if (result == null || result.files.isEmpty || !mounted) return;
    setState(() {
      _pickedFile = result.files.single;
      _pdfError = null;
    });
  }

  void _clearPdf() {
    setState(() => _pickedFile = null);
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    final formValid = _formKey.currentState?.validate() ?? false;
    final pickedFile = _pickedFile;
    if (pickedFile == null && _requestCertification) {
      setState(() => _pdfError = l10n.honeyBatchPdfRequired);
    }
    if (!formValid || (pickedFile == null && _requestCertification)) return;

    setState(() => _saving = true);
    final repo = HoneyBatchRepository(api: context.read<ApiClient>());
    try {
      final amountKg = double.parse(
        _amountController.text.trim().replaceAll(',', '.'),
      );
      await repo.createBatch(
        gatheringDate: _gatheringDate,
        amountGrams: (amountKg * 1000).round(),
        processingMethod: _processingMethod,
        honeyType: _honeyTypeController.text.trim(),
        pdfBytes: pickedFile?.bytes,
        pdfFilename: pickedFile?.name,
        requestCertification: _requestCertification,
      );
      if (!mounted) return;
      Navigator.of(context).pop(true);
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
        setState(() => _saving = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final dateStr = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).format(_gatheringDate);

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.honeyBatchAdd),
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
                                labelText: l10n.honeyBatchGatheringDate,
                              ),
                              child: Text(dateStr),
                            ),
                          ),
                          const SizedBox(height: 16),
                          TextFormField(
                            controller: _amountController,
                            decoration: InputDecoration(
                              labelText: l10n.honeyBatchAmountKg,
                            ),
                            keyboardType: const TextInputType.numberWithOptions(
                              decimal: true,
                            ),
                            textInputAction: TextInputAction.next,
                            autovalidateMode: AutovalidateMode.onUserInteraction,
                            validator: (v) {
                              if (v == null || v.trim().isEmpty) {
                                return l10n.honeyBatchAmountRequired;
                              }
                              final parsed =
                                  double.tryParse(v.trim().replaceAll(',', '.'));
                              if (parsed == null || parsed <= 0) {
                                return l10n.honeyBatchAmountInvalid;
                              }
                              return null;
                            },
                          ),
                          const SizedBox(height: 16),
                          DropdownButtonFormField<ProcessingMethod>(
                            initialValue: _processingMethod,
                            decoration: InputDecoration(
                              labelText: l10n.honeyBatchProcessingMethod,
                            ),
                            items: [
                              for (final method in ProcessingMethod.values)
                                DropdownMenuItem(
                                  value: method,
                                  child: Text(processingMethodLabel(l10n, method)),
                                ),
                            ],
                            onChanged: (value) => setState(
                              () => _processingMethod = value ?? _processingMethod,
                            ),
                          ),
                          const SizedBox(height: 16),
                          TextFormField(
                            controller: _honeyTypeController,
                            decoration: InputDecoration(
                              labelText: l10n.honeyBatchHoneyType,
                              counterText: SizeTier.small.counterText,
                            ),
                            textInputAction: TextInputAction.next,
                            maxLength: SizeTier.small.maxLength,
                            autovalidateMode: AutovalidateMode.onUserInteraction,
                            validator: (v) {
                              if (v == null || v.trim().isEmpty) {
                                return l10n.honeyBatchHoneyTypeRequired;
                              }
                              return validateSizeTier(
                                v,
                                SizeTier.small,
                                l10n.honeyBatchHoneyType,
                                l10n,
                              );
                            },
                          ),
                          const SizedBox(height: 16),
                          _PdfPickerField(
                            file: _pickedFile,
                            error: _pdfError,
                            onPick: _pickPdf,
                            onClear: _clearPdf,
                          ),
                          const SizedBox(height: 16),
                          SwitchListTile(
                            contentPadding: EdgeInsets.zero,
                            title: Text(l10n.honeyBatchCertifyToggle),
                            value: _requestCertification,
                            onChanged: (v) => setState(() {
                              _requestCertification = v;
                              if (!v) _pdfError = null;
                            }),
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
                        _saving
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
                                onPressed: _submit,
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

class _PdfPickerField extends StatelessWidget {
  final PlatformFile? file;
  final String? error;
  final VoidCallback onPick;
  final VoidCallback onClear;

  const _PdfPickerField({
    required this.file,
    required this.error,
    required this.onPick,
    required this.onClear,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: OutlinedButton.icon(
                onPressed: onPick,
                icon: const Icon(Icons.picture_as_pdf_outlined),
                label: Text(
                  file?.name ?? l10n.honeyBatchPdfLabel,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ),
            if (file != null)
              IconButton(
                icon: const Icon(Icons.close),
                onPressed: onClear,
              ),
          ],
        ),
        if (error != null) ...[
          const SizedBox(height: 4),
          Text(error!, style: TextStyle(color: colorScheme.error, fontSize: 12)),
        ],
      ],
    );
  }
}
