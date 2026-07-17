import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../data/feeding_repository.dart';
import 'feeding_form_fields.dart';

class BulkFeedingFormScreen extends StatefulWidget {
  final int apiaryId;
  final List<int>? hiveIds;

  const BulkFeedingFormScreen({super.key, required this.apiaryId, this.hiveIds});

  @override
  State<BulkFeedingFormScreen> createState() => _BulkFeedingFormScreenState();
}

class _BulkFeedingFormScreenState extends State<BulkFeedingFormScreen> {
  final _formKey = GlobalKey<FormState>();

  late DateTime _fedAt;
  late final TextEditingController _feedTypeController;
  late final TextEditingController _amountController;
  late final TextEditingController _notesController;

  List<String> _feedTypeOptions = [];
  List<String> _amountOptions = [];
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _fedAt = DateTime.now();
    _feedTypeController = TextEditingController();
    _amountController = TextEditingController();
    _notesController = TextEditingController();
    _loadSuggestions();
  }

  Future<void> _loadSuggestions() async {
    final repo = FeedingRepository(api: context.read<ApiClient>());
    try {
      final options = await repo.listFeedTypes();
      if (mounted) setState(() => _feedTypeOptions = options);
    } catch (_) {}
    try {
      final amounts = await repo.listAmounts();
      if (mounted) setState(() => _amountOptions = amounts);
    } catch (_) {}
  }

  @override
  void dispose() {
    _feedTypeController.dispose();
    _amountController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _fedAt,
      firstDate: DateTime(2000),
      lastDate: DateTime.now(),
    );
    if (picked == null || !mounted) return;

    final pickedTime = await showTimePicker(
      context: context,
      initialTime: TimeOfDay.fromDateTime(_fedAt),
    );
    if (!mounted) return;
    setState(
      () => _fedAt = DateTime(
        picked.year,
        picked.month,
        picked.day,
        pickedTime?.hour ?? _fedAt.hour,
        pickedTime?.minute ?? _fedAt.minute,
      ),
    );
  }

  Future<void> _submit(BuildContext ctx) async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() => _loading = true);
    final repo = FeedingRepository(api: ctx.read<ApiClient>());
    try {
      final count = await repo.bulkFeeding(
        apiaryId: widget.apiaryId,
        hiveIds: widget.hiveIds,
        fedAt: _fedAt,
        feedType: _feedTypeController.text.trim(),
        amount: _amountController.text.trim(),
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
        title: Text(l10n.feedingFeedAllHives),
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
                    child: FeedingFormFields(
                      formKey: _formKey,
                      fedAt: _fedAt,
                      feedTypeController: _feedTypeController,
                      amountController: _amountController,
                      notesController: _notesController,
                      feedTypeOptions: _feedTypeOptions,
                      amountOptions: _amountOptions,
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
