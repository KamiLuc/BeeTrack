import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/validation/size_tiers.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../data/honey_batch_model.dart';
import '../data/honey_batch_repository.dart';
import '../data/processing_method.dart';

class CreateHoneyBatchScreen extends StatefulWidget {
  final HoneyBatchModel? existingBatch;

  const CreateHoneyBatchScreen({super.key, this.existingBatch});

  bool get isEditing => existingBatch != null;

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
  bool _removeExistingPdf = false;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    final existing = widget.existingBatch;
    if (existing != null) {
      _gatheringDate = existing.gatheringDate;
      _processingMethod = existing.processingMethod;
      _amountController.text = existing.amountKg.toString();
      _honeyTypeController.text = existing.honeyType;
    }
  }

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
      _removeExistingPdf = false;
    });
  }

  void _clearPdf() {
    setState(() {
      if (_pickedFile != null) {
        _pickedFile = null;
      } else {
        _removeExistingPdf = true;
      }
    });
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    final formValid = _formKey.currentState?.validate() ?? false;
    if (!formValid) return;

    setState(() => _saving = true);
    final repo = HoneyBatchRepository(api: context.read<ApiClient>());
    try {
      final amountKg = double.parse(
        _amountController.text.trim().replaceAll(',', '.'),
      );
      if (widget.isEditing) {
        final updated = await repo.updateBatch(
          id: widget.existingBatch!.id,
          gatheringDate: _gatheringDate,
          amountGrams: (amountKg * 1000).round(),
          processingMethod: _processingMethod,
          honeyType: _honeyTypeController.text.trim(),
          pdfBytes: _pickedFile?.bytes,
          pdfFilename: _pickedFile?.name,
          removePdf: _removeExistingPdf,
        );
        if (!mounted) return;
        Navigator.of(context).pop(updated);
      } else {
        await repo.createBatch(
          gatheringDate: _gatheringDate,
          amountGrams: (amountKg * 1000).round(),
          processingMethod: _processingMethod,
          honeyType: _honeyTypeController.text.trim(),
          pdfBytes: _pickedFile?.bytes,
          pdfFilename: _pickedFile?.name,
        );
        if (!mounted) return;
        Navigator.of(context).pop(true);
      }
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
        title: Text(widget.isEditing ? l10n.honeyBatchEditTitle : l10n.honeyBatchAdd),
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
                            existingFilename: widget.existingBatch?.pdfFilename,
                            removed: _removeExistingPdf,
                            onPick: _pickPdf,
                            onClear: _clearPdf,
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
  final String? existingFilename;
  final bool removed;
  final VoidCallback onPick;
  final VoidCallback onClear;

  const _PdfPickerField({
    required this.file,
    this.existingFilename,
    required this.removed,
    required this.onPick,
    required this.onClear,
  });

  bool get _hasExisting =>
      !removed && existingFilename != null && existingFilename!.isNotEmpty;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final label = file?.name ??
        (_hasExisting ? existingFilename! : l10n.honeyBatchPdfLabel);
    final showClear = file != null || _hasExisting;

    return Row(
      children: [
        Expanded(
          child: OutlinedButton.icon(
            onPressed: onPick,
            icon: const Icon(Icons.picture_as_pdf_outlined),
            label: Text(label, overflow: TextOverflow.ellipsis),
          ),
        ),
        if (showClear)
          IconButton(
            icon: const Icon(Icons.close),
            onPressed: onClear,
          ),
      ],
    );
  }
}
