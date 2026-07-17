import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../../core/validation/size_tiers.dart';
import '../../../l10n/app_localizations.dart';

class FeedingFormFields extends StatelessWidget {
  final GlobalKey<FormState> formKey;
  final DateTime fedAt;
  final TextEditingController feedTypeController;
  final TextEditingController amountController;
  final TextEditingController notesController;
  final List<String> feedTypeOptions;
  final List<String> amountOptions;
  final VoidCallback onDateTap;

  const FeedingFormFields({
    super.key,
    required this.formKey,
    required this.fedAt,
    required this.feedTypeController,
    required this.amountController,
    required this.notesController,
    required this.feedTypeOptions,
    this.amountOptions = const [],
    required this.onDateTap,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final dateStr = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).add_Hm().format(fedAt);

    return Form(
      key: formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InkWell(
            onTap: onDateTap,
            child: InputDecorator(
              decoration: InputDecoration(
                labelText: l10n.feedingDate,
                suffixIcon: const Icon(Icons.calendar_today, size: 20),
              ),
              child: Text(dateStr),
            ),
          ),
          const SizedBox(height: 16),
          Autocomplete<String>(
            initialValue: TextEditingValue(text: feedTypeController.text),
            optionsBuilder: (textEditingValue) {
              if (textEditingValue.text.isEmpty) return feedTypeOptions;
              final query = textEditingValue.text.toLowerCase();
              return feedTypeOptions.where((m) => m.toLowerCase().contains(query));
            },
            onSelected: (selection) {
              feedTypeController.text = selection;
            },
            fieldViewBuilder: (context, controller, focusNode, onFieldSubmitted) {
              if (controller.text.isEmpty && feedTypeController.text.isNotEmpty) {
                controller.text = feedTypeController.text;
              }
              return TextFormField(
                controller: controller,
                focusNode: focusNode,
                decoration: InputDecoration(
                  labelText: l10n.feedingType,
                  counterText: SizeTier.small.counterText,
                ),
                maxLength: SizeTier.small.maxLength,
                onChanged: (v) => feedTypeController.text = v,
                validator: (v) {
                  if (v == null || v.trim().isEmpty) {
                    return l10n.feedingTypeRequired;
                  }
                  return validateSizeTier(
                    v,
                    SizeTier.small,
                    l10n.feedingType,
                    l10n,
                  );
                },
              );
            },
          ),
          const SizedBox(height: 16),
          Autocomplete<String>(
            initialValue: TextEditingValue(text: amountController.text),
            optionsBuilder: (textEditingValue) {
              if (textEditingValue.text.isEmpty) return amountOptions;
              final query = textEditingValue.text.toLowerCase();
              return amountOptions.where((a) => a.toLowerCase().contains(query));
            },
            onSelected: (selection) {
              amountController.text = selection;
            },
            fieldViewBuilder: (context, controller, focusNode, onFieldSubmitted) {
              if (controller.text.isEmpty && amountController.text.isNotEmpty) {
                controller.text = amountController.text;
              }
              return TextFormField(
                controller: controller,
                focusNode: focusNode,
                decoration: InputDecoration(
                  labelText: l10n.feedingAmount,
                  counterText: SizeTier.superSmall.counterText,
                ),
                maxLength: SizeTier.superSmall.maxLength,
                onChanged: (v) => amountController.text = v,
                validator: (v) {
                  if (v == null || v.trim().isEmpty) {
                    return l10n.feedingAmountRequired;
                  }
                  return validateSizeTier(
                    v,
                    SizeTier.superSmall,
                    l10n.feedingAmount,
                    l10n,
                  );
                },
              );
            },
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: notesController,
            decoration: InputDecoration(labelText: l10n.feedingNote),
            maxLines: 3,
            minLines: 1,
            maxLength: SizeTier.extraLarge.maxLength,
            validator: (v) => validateSizeTier(
              v,
              SizeTier.extraLarge,
              l10n.feedingNote,
              l10n,
            ),
          ),
        ],
      ),
    );
  }
}
