class ListingImage {
  final int id;
  final int listingId;
  final String url;
  final int displayOrder;
  final DateTime createdAt;

  const ListingImage({
    required this.id,
    required this.listingId,
    required this.url,
    required this.displayOrder,
    required this.createdAt,
  });

  factory ListingImage.fromJson(Map<String, dynamic> json) => ListingImage(
        id: json['id'] as int,
        listingId: json['listing_id'] as int,
        url: json['url'] as String,
        displayOrder: json['display_order'] as int? ?? 0,
        createdAt: DateTime.parse(json['created_at'] as String).toLocal(),
      );
}

class ListingHoneyBatch {
  final int id;
  final String honeyType;
  final DateTime gatheringDate;
  final int amountGrams;
  final String processingMethod;
  final String certificationStatus;
  final bool hasPdf;
  final String verificationUrl;
  final String? pdfUrl;

  const ListingHoneyBatch({
    required this.id,
    required this.honeyType,
    required this.gatheringDate,
    required this.amountGrams,
    required this.processingMethod,
    required this.certificationStatus,
    required this.hasPdf,
    required this.verificationUrl,
    this.pdfUrl,
  });

  double get amountKg => amountGrams / 1000;

  factory ListingHoneyBatch.fromJson(Map<String, dynamic> json) =>
      ListingHoneyBatch(
        id: json['id'] as int,
        honeyType: json['honey_type'] as String? ?? '',
        gatheringDate:
            DateTime.parse(json['gathering_date'] as String).toLocal(),
        amountGrams: json['amount_grams'] as int? ?? 0,
        processingMethod: json['processing_method'] as String? ?? '',
        certificationStatus: json['certification_status'] as String? ?? '',
        hasPdf: json['has_pdf'] as bool? ?? false,
        verificationUrl: json['verification_url'] as String? ?? '',
        pdfUrl: json['pdf_url'] as String?,
      );
}

enum ListingStatus {
  pending('pending'),
  approved('approved'),
  rejected('rejected'),
  removed('removed');

  final String value;

  const ListingStatus(this.value);

  factory ListingStatus.fromJson(String value) {
    return ListingStatus.values.firstWhere((s) => s.value == value);
  }

  String toJson() => value;
}

class Listing {
  final int id;
  final int userId;
  final String title;
  final String description;
  final String category;
  final double? price;
  final String quantity;
  final String address;
  final double lat;
  final double lng;
  final int? apiaryId;
  final String? apiaryName;
  final double? apiaryLat;
  final double? apiaryLng;
  final int apiaryHiveCount;
  final String contactPhone;
  final String contactEmail;
  final bool isHidden;
  final ListingStatus status;
  final String? rejectionReason;
  final DateTime createdAt;
  final DateTime updatedAt;
  final List<ListingImage> images;
  final double? distanceKm;
  final int? honeyBatchId;
  final ListingHoneyBatch? honeyBatch;

  const Listing({
    required this.id,
    required this.userId,
    required this.title,
    required this.description,
    required this.category,
    this.price,
    required this.quantity,
    required this.address,
    required this.lat,
    required this.lng,
    this.apiaryId,
    this.apiaryName,
    this.apiaryLat,
    this.apiaryLng,
    this.apiaryHiveCount = 0,
    required this.contactPhone,
    required this.contactEmail,
    required this.isHidden,
    this.status = ListingStatus.approved,
    this.rejectionReason,
    required this.createdAt,
    required this.updatedAt,
    required this.images,
    this.distanceKm,
    this.honeyBatchId,
    this.honeyBatch,
  });

  factory Listing.fromJson(Map<String, dynamic> json) => Listing(
        id: json['id'] as int,
        userId: json['user_id'] as int,
        title: json['title'] as String,
        description: json['description'] as String? ?? '',
        category: json['category'] as String,
        price: (json['price'] as num?)?.toDouble(),
        quantity: json['quantity'] as String? ?? '',
        address: json['address'] as String? ?? '',
        lat: (json['lat'] as num?)?.toDouble() ?? 0,
        lng: (json['lng'] as num?)?.toDouble() ?? 0,
        apiaryId: json['apiary_id'] as int?,
        apiaryName: json['apiary_name'] as String?,
        apiaryLat: (json['apiary_lat'] as num?)?.toDouble(),
        apiaryLng: (json['apiary_lng'] as num?)?.toDouble(),
        apiaryHiveCount: json['apiary_hive_count'] as int? ?? 0,
        contactPhone: json['contact_phone'] as String? ?? '',
        contactEmail: json['contact_email'] as String? ?? '',
        isHidden: json['is_hidden'] as bool? ?? false,
        status: json['status'] != null
            ? ListingStatus.fromJson(json['status'] as String)
            : ListingStatus.approved,
        rejectionReason: json['rejection_reason'] as String?,
        createdAt: DateTime.parse(json['created_at'] as String).toLocal(),
        updatedAt: DateTime.parse(json['updated_at'] as String).toLocal(),
        images: (json['images'] as List<dynamic>? ?? [])
            .map((e) => ListingImage.fromJson(e as Map<String, dynamic>))
            .toList(),
        distanceKm: (json['distance_km'] as num?)?.toDouble(),
        honeyBatchId: json['honey_batch_id'] as int?,
        honeyBatch: json['honey_batch'] != null
            ? ListingHoneyBatch.fromJson(
                json['honey_batch'] as Map<String, dynamic>)
            : null,
      );
}
