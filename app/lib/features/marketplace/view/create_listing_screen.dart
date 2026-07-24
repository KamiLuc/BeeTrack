import 'dart:typed_data';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:geolocator/geolocator.dart';
import 'package:image_picker/image_picker.dart';
import 'package:intl/intl.dart';
import 'package:latlong2/latlong.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/validation/gps_bounds.dart';
import '../../../core/validation/size_tiers.dart';
import '../../../core/widgets/location_picker_section.dart';
import '../../../core/widgets/map_picker_screen.dart';
import '../../../core/widgets/photo_size_snackbar.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../features/apiary/data/apiary_model.dart';
import '../../../features/apiary/data/apiary_repository.dart';
import '../../../features/honey_batch/data/honey_batch_certification_model.dart';
import '../../../features/honey_batch/data/honey_batch_model.dart';
import '../../../features/honey_batch/data/honey_batch_repository.dart';
import '../../../features/honey_batch/data/processing_method.dart';
import '../../../l10n/app_localizations.dart';
import '../data/listing_category.dart';
import '../data/listing_model.dart';
import '../data/listing_repository.dart';

const int _kMaxListingPhotos = 3;

class CreateListingScreen extends StatefulWidget {
  final Listing? existingListing;

  const CreateListingScreen({super.key, this.existingListing});

  bool get isEditing => existingListing != null;

  @override
  State<CreateListingScreen> createState() => _CreateListingScreenState();
}

class _CreateListingScreenState extends State<CreateListingScreen> {
  final _formKey = GlobalKey<FormState>();
  final _titleController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _priceController = TextEditingController();
  final _quantityController = TextEditingController();
  final _addressController = TextEditingController();
  final _latController = TextEditingController();
  final _lngController = TextEditingController();
  final _phoneController = TextEditingController();
  final _emailController = TextEditingController();

  String? _category;
  int? _apiaryId;
  List<Apiary> _apiaries = [];
  int? _honeyBatchId;
  List<HoneyBatchModel> _certifiedBatches = [];
  bool _loadingBatches = false;
  final List<XFile> _pendingImages = [];
  late List<ListingImage> _existingImages;
  final Set<int> _pendingRemovalIds = {};
  final Map<XFile, double> _uploadProgress = {};
  bool _saving = false;
  bool _locating = false;
  String? _submitError;

  @override
  void initState() {
    super.initState();
    _loadApiaries();
    final listing = widget.existingListing;
    _existingImages = List.of(listing?.images ?? const []);
    if (listing != null) {
      _titleController.text = listing.title;
      _descriptionController.text = listing.description;
      _priceController.text = listing.price != null
          ? listing.price!.toStringAsFixed(2)
          : '';
      _quantityController.text = listing.quantity;
      _addressController.text = listing.address;
      _setLocation(LatLng(listing.lat, listing.lng));
      _phoneController.text = listing.contactPhone;
      _emailController.text = listing.contactEmail;
      _category = listing.category;
      _apiaryId = listing.apiaryId;
      _honeyBatchId = listing.honeyBatchId;
    }
    if (_category == 'HONEY') _loadCertifiedBatches();
  }

  @override
  void dispose() {
    _titleController.dispose();
    _descriptionController.dispose();
    _priceController.dispose();
    _quantityController.dispose();
    _addressController.dispose();
    _latController.dispose();
    _lngController.dispose();
    _phoneController.dispose();
    _emailController.dispose();
    super.dispose();
  }

  LatLng? get _location {
    final lat = double.tryParse(_latController.text);
    final lng = double.tryParse(_lngController.text);
    if (lat != null && lng != null) return LatLng(lat, lng);
    return null;
  }

  void _setLocation(LatLng loc) {
    _latController.text = clampLatitude(loc.latitude).toStringAsFixed(6);
    _lngController.text = clampLongitude(loc.longitude).toStringAsFixed(6);
  }

  Future<void> _useGps() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() => _locating = true);
    try {
      final serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        _showGpsError(l10n);
        return;
      }
      var permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
      }
      if (permission == LocationPermission.denied ||
          permission == LocationPermission.deniedForever) {
        _showGpsError(l10n);
        return;
      }
      final pos = await Geolocator.getCurrentPosition();
      setState(() => _setLocation(LatLng(pos.latitude, pos.longitude)));
      _reviewLocationError();
    } catch (_) {
      _showGpsError(l10n);
    } finally {
      if (mounted) setState(() => _locating = false);
    }
  }

  void _showGpsError(AppLocalizations l10n) {
    if (!mounted) return;
    showBigSnackBar(context, l10n.marketplaceGpsUnavailable);
  }

  Future<void> _pickOnMap() async {
    final result = await Navigator.of(context).push<LatLng>(
      MaterialPageRoute(builder: (_) => MapPickerScreen(initial: _location)),
    );
    if (result != null) {
      setState(() => _setLocation(result));
      _reviewLocationError();
    }
  }

  void _reviewLocationError() {
    if (_submitError == null) return;
    if (_submitError !=
        AppLocalizations.of(context)!.marketplaceLocationRequired) {
      return;
    }
    if (_location != null) setState(() => _submitError = null);
  }

  Future<void> _loadApiaries() async {
    try {
      final apiaries = await ApiaryRepository(
        api: context.read<ApiClient>(),
      ).listApiaries();
      if (mounted) setState(() => _apiaries = apiaries);
    } catch (_) {}
  }

  /// Loads the caller's own honey batches with a confirmed on-chain
  /// certification — only these can be attached to a HONEY listing.
  Future<void> _loadCertifiedBatches() async {
    setState(() => _loadingBatches = true);
    try {
      final result = await HoneyBatchRepository(
        api: context.read<ApiClient>(),
      ).listBatches(limit: 100);
      final confirmed = result.items
          .where(
            (b) => b.certification?.status == CertificationStatus.confirmed,
          )
          .toList();
      if (mounted) {
        setState(() {
          _certifiedBatches = confirmed;
          _loadingBatches = false;
        });
      }
    } catch (_) {
      if (mounted) setState(() => _loadingBatches = false);
    }
  }

  void _onCategoryChanged(String? value) {
    setState(() {
      _category = value;
      if (value != 'HONEY') _honeyBatchId = null;
    });
    if (value == 'HONEY' && _certifiedBatches.isEmpty && !_loadingBatches) {
      _loadCertifiedBatches();
    }
  }

  bool _isValidEmail(String v) {
    return RegExp(
      r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$',
    ).hasMatch(v.trim());
  }

  bool _isValidPhone(String v) {
    final digits = v.replaceAll(RegExp(r'[^0-9]'), '');
    return digits.length >= 9;
  }

  static final _priceInputFormatter = TextInputFormatter.withFunction((
    oldValue,
    newValue,
  ) {
    if (newValue.text.isEmpty) return newValue;
    return RegExp(r'^\d*\.?\d{0,2}$').hasMatch(newValue.text)
        ? newValue
        : oldValue;
  });

  int get _totalPhotoCount => _existingImages.length + _pendingImages.length;

  Future<void> _pickImage() async {
    if (_totalPhotoCount >= _kMaxListingPhotos) return;
    final l10n = AppLocalizations.of(context)!;
    final picker = ImagePicker();
    final ImageSource source;
    if (kIsWeb) {
      source = ImageSource.gallery;
    } else {
      final chosen = await showModalBottomSheet<ImageSource>(
        context: context,
        builder: (_) => SafeArea(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              ListTile(
                leading: const Icon(Icons.photo_library_outlined),
                title: Text(l10n.marketplacePhotoSourceGallery),
                onTap: () => Navigator.of(context).pop(ImageSource.gallery),
              ),
              ListTile(
                leading: const Icon(Icons.camera_alt_outlined),
                title: Text(l10n.marketplacePhotoSourceCamera),
                onTap: () => Navigator.of(context).pop(ImageSource.camera),
              ),
            ],
          ),
        ),
      );
      if (chosen == null) return;
      source = chosen;
    }
    final file = await picker.pickImage(source: source, imageQuality: 85);
    if (file == null || !mounted) return;
    final sizeError = await validateImageFileSize(file, l10n);
    if (sizeError != null) {
      if (mounted) showBigSnackBar(context, sizeError);
      return;
    }
    if (mounted) setState(() => _pendingImages.add(file));
    _reviewPhotoError();
  }

  void _removeImage(XFile file) {
    setState(() => _pendingImages.remove(file));
  }

  /// Removes an already-uploaded photo locally — the deletion itself is
  /// deferred to submit, so removing a listing's only photo can still be
  /// undone (or flagged by the photo-required check) before it's persisted.
  void _removeExistingImage(ListingImage image) {
    setState(() {
      _existingImages.remove(image);
      _pendingRemovalIds.add(image.id);
    });
    _reviewPhotoError();
  }

  void _reviewPhotoError() {
    if (_submitError == null) return;
    if (_submitError != AppLocalizations.of(context)!.marketplacePhotoRequired) {
      return;
    }
    if (_totalPhotoCount > 0) setState(() => _submitError = null);
  }

  void _reviewContactError() {
    if (_submitError == null) return;
    if (_submitError !=
        AppLocalizations.of(context)!.marketplaceContactRequired) {
      return;
    }
    final hasContact =
        _phoneController.text.trim().isNotEmpty ||
        _emailController.text.trim().isNotEmpty;
    if (hasContact) setState(() => _submitError = null);
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    final formValid = _formKey.currentState?.validate() ?? false;
    final hasContact =
        _phoneController.text.trim().isNotEmpty ||
        _emailController.text.trim().isNotEmpty;
    final location = _location;
    final hasPhoto = _totalPhotoCount > 0;
    if (!formValid ||
        _category == null ||
        !hasContact ||
        location == null ||
        !hasPhoto) {
      setState(() {
        _submitError = !hasContact
            ? l10n.marketplaceContactRequired
            : location == null
            ? l10n.marketplaceLocationRequired
            : !hasPhoto
            ? l10n.marketplacePhotoRequired
            : null;
      });
      return;
    }

    setState(() {
      _submitError = null;
      _saving = true;
    });
    final repo = ListingRepository(api: context.read<ApiClient>());
    try {
      final int listingId;
      if (widget.isEditing) {
        await repo.updateListing(
          id: widget.existingListing!.id,
          title: _titleController.text.trim(),
          description: _descriptionController.text.trim(),
          category: _category!,
          price: double.parse(_priceController.text.trim()),
          quantity: _quantityController.text.trim(),
          address: _addressController.text.trim(),
          lat: location.latitude,
          lng: location.longitude,
          apiaryId: _apiaryId,
          contactPhone: _phoneController.text.trim(),
          contactEmail: _emailController.text.trim(),
          honeyBatchId: _honeyBatchId,
        );
        listingId = widget.existingListing!.id;
      } else {
        final listing = await repo.createListing(
          title: _titleController.text.trim(),
          description: _descriptionController.text.trim(),
          category: _category!,
          price: double.parse(_priceController.text.trim()),
          quantity: _quantityController.text.trim(),
          address: _addressController.text.trim(),
          lat: location.latitude,
          lng: location.longitude,
          apiaryId: _apiaryId,
          contactPhone: _phoneController.text.trim(),
          contactEmail: _emailController.text.trim(),
          honeyBatchId: _honeyBatchId,
        );
        listingId = listing.id;
      }
      // Walk the queued removals/uploads one step at a time, tracking the
      // photo count the server actually has right now (it starts at the
      // pre-edit count, since none of today's deletes/uploads have run yet).
      // At each step we only delete while that count is still above 1, and
      // only upload while it's still below the 3-photo cap — both of which
      // the backend enforces — so e.g. swapping a listing's only photo for
      // a new one uploads the replacement first instead of deleting the
      // last photo out from under an empty listing.
      var serverPhotoCount = widget.existingListing?.images.length ?? 0;
      while (_pendingRemovalIds.isNotEmpty || _pendingImages.isNotEmpty) {
        if (_pendingRemovalIds.isNotEmpty && serverPhotoCount > 1) {
          final id = _pendingRemovalIds.first;
          await repo.deleteImage(listingId, id);
          if (mounted) setState(() => _pendingRemovalIds.remove(id));
          serverPhotoCount--;
        } else if (_pendingImages.isNotEmpty &&
            serverPhotoCount < _kMaxListingPhotos) {
          final file = _pendingImages.first;
          await repo.uploadImage(
            listingId,
            file,
            onSendProgress: (sent, total) {
              if (total <= 0 || !mounted) return;
              setState(() => _uploadProgress[file] = sent / total);
            },
          );
          if (mounted) setState(() => _pendingImages.remove(file));
          serverPhotoCount++;
        } else {
          break;
        }
      }
      if (mounted) Navigator.of(context).pop(true);
    } catch (e) {
      if (mounted) {
        setState(() {
          _submitError = switch ((e is ApiException) ? e.code : null) {
            'LISTING_LIMIT_REACHED' => l10n.marketplaceListingLimitReached,
            'LAST_PHOTO' => l10n.marketplaceLastPhotoRequired,
            _ => l10n.generalError,
          };
          _saving = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final textTheme = Theme.of(context).textTheme;
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: Text(
          widget.isEditing
              ? l10n.marketplaceEditScreenTitle
              : l10n.marketplaceCreateScreenTitle,
        ),
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
                          TextFormField(
                            controller: _titleController,
                            decoration: InputDecoration(
                              labelText: l10n.marketplaceFieldTitle,
                              counterText: SizeTier.medium.counterText,
                            ),
                            textInputAction: TextInputAction.next,
                            maxLength: SizeTier.medium.maxLength,
                            autovalidateMode:
                                AutovalidateMode.onUserInteraction,
                            validator: (v) {
                              if (v == null || v.trim().isEmpty) {
                                return l10n.marketplaceFieldTitleRequired;
                              }
                              return validateSizeTier(
                                v,
                                SizeTier.medium,
                                l10n.marketplaceFieldTitle,
                                l10n,
                              );
                            },
                          ),
                          const SizedBox(height: 16),
                          DropdownButtonFormField<String>(
                            initialValue: _category,
                            decoration: InputDecoration(
                              labelText: l10n.marketplaceFieldCategory,
                            ),
                            autovalidateMode:
                                AutovalidateMode.onUserInteraction,
                            items: [
                              for (final category in listingCategories)
                                DropdownMenuItem(
                                  value: category,
                                  child: Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Icon(
                                        listingCategoryIcon(category),
                                        size: 20,
                                      ),
                                      const SizedBox(width: 8),
                                      Text(
                                        listingCategoryLabel(l10n, category),
                                      ),
                                    ],
                                  ),
                                ),
                            ],
                            onChanged: _onCategoryChanged,
                            validator: (v) => v == null
                                ? l10n.marketplaceFieldCategoryRequired
                                : null,
                          ),
                          if (_category == 'HONEY') ...[
                            const SizedBox(height: 16),
                            _HoneyBatchAttachSection(
                              loading: _loadingBatches,
                              batches: _certifiedBatches,
                              selectedId: _honeyBatchId,
                              onChanged: (value) =>
                                  setState(() => _honeyBatchId = value),
                            ),
                          ],
                          const SizedBox(height: 16),
                          TextFormField(
                            controller: _descriptionController,
                            decoration: InputDecoration(
                              labelText: l10n.marketplaceDescriptionLabel,
                            ),
                            minLines: 3,
                            maxLines: 6,
                            maxLength: SizeTier.large.maxLength,
                            textInputAction: TextInputAction.newline,
                            autovalidateMode:
                                AutovalidateMode.onUserInteraction,
                            validator: (v) => validateSizeTier(
                              v,
                              SizeTier.large,
                              l10n.marketplaceDescriptionLabel,
                              l10n,
                            ),
                          ),
                          const SizedBox(height: 16),
                          TextFormField(
                            controller: _priceController,
                            decoration: InputDecoration(
                              labelText: l10n.marketplaceFieldPrice,
                              counterText: SizeTier.small.counterText,
                            ),
                            keyboardType: TextInputType.numberWithOptions(
                              decimal: true,
                            ),
                            textInputAction: TextInputAction.next,
                            inputFormatters: [_priceInputFormatter],
                            maxLength: SizeTier.small.maxLength,
                            autovalidateMode:
                                AutovalidateMode.onUserInteraction,
                            validator: (v) {
                              if (v == null || v.trim().isEmpty) {
                                return l10n.marketplaceFieldPriceRequired;
                              }
                              final parsed = double.tryParse(v.trim());
                              if (parsed == null) {
                                return l10n.marketplaceFieldPriceInvalid;
                              }
                              if (parsed >= 100000000) {
                                return l10n.marketplaceFieldPriceTooLarge;
                              }
                              return validateSizeTier(
                                v,
                                SizeTier.small,
                                l10n.marketplaceFieldPrice,
                                l10n,
                              );
                            },
                          ),
                          const SizedBox(height: 16),
                          TextFormField(
                            controller: _quantityController,
                            decoration: InputDecoration(
                              labelText: l10n.marketplaceQuantityLabel,
                              counterText: SizeTier.small.counterText,
                            ),
                            textInputAction: TextInputAction.next,
                            maxLength: SizeTier.small.maxLength,
                            autovalidateMode:
                                AutovalidateMode.onUserInteraction,
                            validator: (v) => validateSizeTier(
                              v,
                              SizeTier.small,
                              l10n.marketplaceQuantityLabel,
                              l10n,
                            ),
                          ),
                          const SizedBox(height: 16),
                          TextFormField(
                            controller: _addressController,
                            decoration: InputDecoration(
                              labelText: l10n.marketplaceFieldAddress,
                              counterText: SizeTier.medium.counterText,
                            ),
                            textInputAction: TextInputAction.next,
                            maxLength: SizeTier.medium.maxLength,
                            autovalidateMode:
                                AutovalidateMode.onUserInteraction,
                            validator: (v) => validateSizeTier(
                              v,
                              SizeTier.medium,
                              l10n.marketplaceFieldAddress,
                              l10n,
                            ),
                          ),
                          const SizedBox(height: 24),
                          LocationPickerSection(
                            latController: _latController,
                            lngController: _lngController,
                            latLabel: l10n.marketplaceFieldLatitude,
                            lngLabel: l10n.marketplaceFieldLongitude,
                            locating: _locating,
                            onGps: _useGps,
                            onMap: _pickOnMap,
                          ),
                          const SizedBox(height: 24),
                          Text(
                            l10n.marketplaceContactLabel,
                            style: textTheme.titleSmall?.copyWith(
                              color: colorScheme.primary,
                            ),
                          ),
                          const SizedBox(height: 8),
                          TextFormField(
                            controller: _phoneController,
                            decoration: InputDecoration(
                              labelText: l10n.marketplaceFieldPhone,
                              counterText: SizeTier.superSmall.counterText,
                            ),
                            keyboardType: TextInputType.phone,
                            textInputAction: TextInputAction.next,
                            maxLength: SizeTier.superSmall.maxLength,
                            autovalidateMode:
                                AutovalidateMode.onUserInteraction,
                            onChanged: (_) => _reviewContactError(),
                            validator: (v) {
                              if (v == null || v.trim().isEmpty) return null;
                              if (!_isValidPhone(v)) {
                                return l10n.marketplaceFieldPhoneInvalid;
                              }
                              return validateSizeTier(
                                v,
                                SizeTier.superSmall,
                                l10n.marketplaceFieldPhone,
                                l10n,
                              );
                            },
                          ),
                          const SizedBox(height: 16),
                          TextFormField(
                            controller: _emailController,
                            decoration: InputDecoration(
                              labelText: l10n.marketplaceFieldEmail,
                              counterText: SizeTier.medium.counterText,
                            ),
                            keyboardType: TextInputType.emailAddress,
                            textInputAction: TextInputAction.next,
                            maxLength: SizeTier.medium.maxLength,
                            autovalidateMode:
                                AutovalidateMode.onUserInteraction,
                            onChanged: (_) => _reviewContactError(),
                            validator: (v) {
                              if (v == null || v.trim().isEmpty) return null;
                              if (!_isValidEmail(v))
                                return l10n.authInvalidEmail;
                              return validateSizeTier(
                                v,
                                SizeTier.medium,
                                l10n.marketplaceFieldEmail,
                                l10n,
                              );
                            },
                          ),
                          if (_apiaries.isNotEmpty) ...[
                            const SizedBox(height: 24),
                            Text(
                              l10n.marketplaceApiaryLabel,
                              style: textTheme.titleSmall?.copyWith(
                                color: colorScheme.primary,
                              ),
                            ),
                            const SizedBox(height: 8),
                            DropdownButtonFormField<int?>(
                              initialValue: _apiaryId,
                              items: [
                                DropdownMenuItem(
                                  value: null,
                                  child: Text(l10n.marketplaceApiaryNone),
                                ),
                                for (final apiary in _apiaries)
                                  DropdownMenuItem(
                                    value: apiary.id,
                                    child: Text(apiary.name),
                                  ),
                              ],
                              onChanged: (value) =>
                                  setState(() => _apiaryId = value),
                            ),
                          ],
                          const SizedBox(height: 24),
                          _PhotosSection(
                            existingImages: _existingImages,
                            pendingImages: _pendingImages,
                            uploadProgress: _uploadProgress,
                            totalCount: _totalPhotoCount,
                            saving: _saving,
                            onRemove: _removeImage,
                            onRemoveExisting: _removeExistingImage,
                          ),
                          if (_submitError != null) ...[
                            const SizedBox(height: 16),
                            Text(
                              _submitError!,
                              textAlign: TextAlign.center,
                              style: TextStyle(color: colorScheme.error),
                            ),
                          ],
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
          _ListingFormBanner(
            loading: _saving,
            pendingCount: _pendingImages.length,
            atPhotoLimit: _totalPhotoCount >= _kMaxListingPhotos,
            onSave: _submit,
            onAddPhoto: _pickImage,
          ),
        ],
      ),
    );
  }
}

class _HoneyBatchAttachSection extends StatelessWidget {
  final bool loading;
  final List<HoneyBatchModel> batches;
  final int? selectedId;
  final ValueChanged<int?> onChanged;

  const _HoneyBatchAttachSection({
    required this.loading,
    required this.batches,
    required this.selectedId,
    required this.onChanged,
  });

  String _batchLabel(BuildContext context, HoneyBatchModel batch) {
    final l10n = AppLocalizations.of(context)!;
    final dateLabel = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).format(batch.gatheringDate);
    final methodLabel = processingMethodLabel(l10n, batch.processingMethod);
    return '${batch.honeyType} · ${batch.amountKg.toStringAsFixed(1)} kg · '
        '$methodLabel · $dateLabel';
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;

    if (!loading && batches.isEmpty) return const SizedBox.shrink();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.marketplaceHoneyBatchAttachLabel,
          style: Theme.of(
            context,
          ).textTheme.titleSmall?.copyWith(color: colorScheme.primary),
        ),
        const SizedBox(height: 8),
        if (loading)
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 8),
            child: SizedBox(
              height: 20,
              width: 20,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
          )
        else
          DropdownButtonFormField<int?>(
            initialValue: selectedId,
            items: [
              DropdownMenuItem(
                value: null,
                child: Text(l10n.marketplaceHoneyBatchNone),
              ),
              for (final batch in batches)
                DropdownMenuItem(
                  value: batch.id,
                  child: Text(
                    _batchLabel(context, batch),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
            ],
            onChanged: onChanged,
          ),
      ],
    );
  }
}

// ── Amber banner: save + add-photo buttons ────────────────────────────────────

class _ListingFormBanner extends StatelessWidget {
  final bool loading;
  final int pendingCount;
  final bool atPhotoLimit;
  final VoidCallback onSave;
  final VoidCallback onAddPhoto;

  const _ListingFormBanner({
    required this.loading,
    required this.pendingCount,
    required this.atPhotoLimit,
    required this.onSave,
    required this.onAddPhoto,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final bannerWidth = AppLayout.bannerWidth(context);

    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.only(bottom: 8),
        child: Center(
          child: SizedBox(
            width: bannerWidth,
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
                  Badge(
                    isLabelVisible: pendingCount > 0,
                    label: Text('$pendingCount'),
                    child: IconButton(
                      icon: const Icon(Icons.add_photo_alternate_outlined),
                      iconSize: 28,
                      tooltip: l10n.marketplaceAddPhoto,
                      onPressed: (atPhotoLimit || loading) ? null : onAddPhoto,
                    ),
                  ),
                  loading
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
                          onPressed: onSave,
                        ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _PhotosSection extends StatelessWidget {
  final List<ListingImage> existingImages;
  final List<XFile> pendingImages;
  final Map<XFile, double> uploadProgress;
  final int totalCount;
  final bool saving;
  final ValueChanged<XFile> onRemove;
  final ValueChanged<ListingImage> onRemoveExisting;

  const _PhotosSection({
    required this.existingImages,
    required this.pendingImages,
    required this.uploadProgress,
    required this.totalCount,
    required this.saving,
    required this.onRemove,
    required this.onRemoveExisting,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final baseUrl = context.read<ApiClient>().baseUrl;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (totalCount > 0) ...[
          Text(
            '${l10n.marketplacePhotosLabel}  $totalCount/$_kMaxListingPhotos',
            style: Theme.of(context).textTheme.titleSmall?.copyWith(
              color: Theme.of(context).colorScheme.primary,
            ),
          ),
          const SizedBox(height: 8),
          LayoutBuilder(
            builder: (context, constraints) {
              final size = kIsWeb ? 180.0 : 120.0;
              return SizedBox(
                height: size + 15,
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: ConstrainedBox(
                    constraints: BoxConstraints(minWidth: constraints.maxWidth),
                    child: Center(
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          for (final image in existingImages)
                            _ExistingThumbnail(
                              image: image,
                              baseUrl: baseUrl,
                              size: size,
                              disabled: saving,
                              onRemove: () => onRemoveExisting(image),
                            ),
                          for (final file in pendingImages)
                            _PendingThumbnail(
                              file: file,
                              size: size,
                              progress: saving
                                  ? uploadProgress[file] ?? 0
                                  : null,
                              onRemove: () => onRemove(file),
                            ),
                        ],
                      ),
                    ),
                  ),
                ),
              );
            },
          ),
        ],
      ],
    );
  }
}

class _ExistingThumbnail extends StatelessWidget {
  final ListingImage image;
  final String baseUrl;
  final double size;
  final bool disabled;
  final VoidCallback onRemove;

  const _ExistingThumbnail({
    required this.image,
    required this.baseUrl,
    required this.size,
    required this.disabled,
    required this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: 8, bottom: 4),
      child: Stack(
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(8),
            child: SizedBox(
              width: size,
              height: size,
              child: Image.network(
                '$baseUrl${image.url}',
                fit: BoxFit.cover,
                errorBuilder: (context, error, stack) => Container(
                  color: Theme.of(context).colorScheme.surfaceContainerHighest,
                  alignment: Alignment.center,
                  child: const Icon(Icons.broken_image_outlined, size: 24),
                ),
              ),
            ),
          ),
          if (!disabled)
            Positioned(
              top: 2,
              right: 2,
              child: GestureDetector(
                onTap: onRemove,
                child: Container(
                  decoration: const BoxDecoration(
                    color: Colors.black54,
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(Icons.close, color: Colors.white, size: 18),
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class _PendingThumbnail extends StatelessWidget {
  final XFile file;
  final double size;

  /// Upload progress in [0, 1] while the listing is being saved, or null
  /// before submission (shows the remove button instead).
  final double? progress;
  final VoidCallback onRemove;

  const _PendingThumbnail({
    required this.file,
    required this.size,
    required this.progress,
    required this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: 8, bottom: 4),
      child: Stack(
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(8),
            child: SizedBox(
              width: size,
              height: size,
              child: FutureBuilder<Uint8List>(
                future: file.readAsBytes(),
                builder: (_, snapshot) {
                  if (snapshot.hasData) {
                    return Image.memory(snapshot.data!, fit: BoxFit.cover);
                  }
                  return const Center(
                    child: CircularProgressIndicator(strokeWidth: 2),
                  );
                },
              ),
            ),
          ),
          if (progress != null)
            Positioned.fill(
              child: ColoredBox(
                color: Colors.black38,
                child: Center(
                  child: SizedBox(
                    width: 28,
                    height: 28,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      value: progress! > 0 ? progress : null,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
            )
          else
            Positioned(
              top: 2,
              right: 2,
              child: GestureDetector(
                onTap: onRemove,
                child: Container(
                  decoration: const BoxDecoration(
                    color: Colors.black54,
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(Icons.close, color: Colors.white, size: 18),
                ),
              ),
            ),
        ],
      ),
    );
  }
}
