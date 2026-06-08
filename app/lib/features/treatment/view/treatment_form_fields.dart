import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';

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
                decoration: InputDecoration(labelText: l10n.treatmentMedicine),
                onChanged: (v) => medicineController.text = v,
                validator: (v) => (v == null || v.trim().isEmpty)
                    ? l10n.treatmentMedicineRequired
                    : null,
              );
            },
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: doseController,
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
            controller: notesController,
            decoration: InputDecoration(labelText: l10n.treatmentNote),
            maxLines: 3,
            minLines: 1,
          ),
        ],
      ),
    );
  }
}
