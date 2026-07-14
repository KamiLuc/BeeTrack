import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';

import '../../../core/validation/size_tiers.dart';
import '../../../l10n/app_localizations.dart';

class TreatmentFormFields extends StatelessWidget {
  final GlobalKey<FormState> formKey;
  final DateTime treatedAt;
  final TextEditingController medicineController;
  final TextEditingController doseController;
  final TextEditingController notesController;
  final List<String> medicineOptions;
  final VoidCallback onDateTap;

  const TreatmentFormFields({
    super.key,
    required this.formKey,
    required this.treatedAt,
    required this.medicineController,
    required this.doseController,
    required this.notesController,
    required this.medicineOptions,
    required this.onDateTap,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final dateStr = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).add_Hm().format(treatedAt);

    return Form(
      key: formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InkWell(
            onTap: onDateTap,
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
            initialValue: TextEditingValue(text: medicineController.text),
            optionsBuilder: (textEditingValue) {
              if (textEditingValue.text.isEmpty) return medicineOptions;
              final query = textEditingValue.text.toLowerCase();
              return medicineOptions.where((m) => m.toLowerCase().contains(query));
            },
            onSelected: (selection) {
              medicineController.text = selection;
            },
            fieldViewBuilder: (context, controller, focusNode, onFieldSubmitted) {
              if (controller.text.isEmpty && medicineController.text.isNotEmpty) {
                controller.text = medicineController.text;
              }
              return TextFormField(
                controller: controller,
                focusNode: focusNode,
                decoration: InputDecoration(
                  labelText: l10n.treatmentMedicine,
                  counterText: SizeTier.small.counterText,
                ),
                maxLength: SizeTier.small.maxLength,
                onChanged: (v) => medicineController.text = v,
                validator: (v) {
                  if (v == null || v.trim().isEmpty) {
                    return l10n.treatmentMedicineRequired;
                  }
                  return validateSizeTier(
                    v,
                    SizeTier.small,
                    l10n.treatmentMedicine,
                    l10n,
                  );
                },
              );
            },
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: doseController,
            decoration: InputDecoration(
              labelText: l10n.treatmentDose,
              counterText: SizeTier.superSmall.counterText,
            ),
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            inputFormatters: [
              FilteringTextInputFormatter.allow(RegExp(r'[0-9.,]')),
            ],
            maxLength: SizeTier.superSmall.maxLength,
            validator: (v) {
              if (v == null || v.trim().isEmpty) {
                return l10n.treatmentDoseRequired;
              }
              return validateSizeTier(
                v,
                SizeTier.superSmall,
                l10n.treatmentDose,
                l10n,
              );
            },
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: notesController,
            decoration: InputDecoration(labelText: l10n.treatmentNote),
            maxLines: 3,
            minLines: 1,
            maxLength: SizeTier.extraLarge.maxLength,
            validator: (v) => validateSizeTier(
              v,
              SizeTier.extraLarge,
              l10n.treatmentNote,
              l10n,
            ),
          ),
        ],
      ),
    );
  }
}
