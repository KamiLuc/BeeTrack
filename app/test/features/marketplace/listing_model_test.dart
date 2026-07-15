import 'package:flutter_test/flutter_test.dart';
import 'package:app/features/marketplace/data/listing_model.dart';

void main() {
  group('ListingImage.fromJson', () {
    test('parses all fields', () {
      final json = {
        'id': 5,
        'listing_id': 3,
        'url': '/api/v1/listings/3/images/5/file',
        'display_order': 2,
        'created_at': '2024-01-15T10:00:00Z',
      };
      final img = ListingImage.fromJson(json);
      expect(img.id, 5);
      expect(img.listingId, 3);
      expect(img.url, '/api/v1/listings/3/images/5/file');
      expect(img.displayOrder, 2);
      expect(img.createdAt, DateTime.parse('2024-01-15T10:00:00Z').toLocal());
    });

    test('defaults display_order to 0 when missing', () {
      final json = {
        'id': 6,
        'listing_id': 3,
        'url': '/api/v1/listings/3/images/6/file',
        'created_at': '2024-01-16T10:00:00Z',
      };
      final img = ListingImage.fromJson(json);
      expect(img.displayOrder, 0);
    });
  });

  group('Listing.fromJson', () {
    test('parses all fields including nested images', () {
      final json = {
        'id': 1,
        'user_id': 42,
        'title': 'Fresh honey',
        'description': '5kg jars',
        'category': 'HONEY',
        'price': 12.5,
        'quantity': '10 jars',
        'address': 'Warsaw',
        'lat': 52.2297,
        'lng': 21.0122,
        'apiary_id': 7,
        'apiary_name': 'Home Apiary',
        'contact_phone': '123456789',
        'contact_email': 'seller@example.com',
        'is_hidden': false,
        'created_at': '2024-02-01T08:30:00Z',
        'updated_at': '2024-02-02T09:00:00Z',
        'distance_km': 3.25,
        'images': [
          {
            'id': 1,
            'listing_id': 1,
            'url': '/api/v1/listings/1/images/1/file',
            'display_order': 0,
            'created_at': '2024-02-01T08:30:00Z',
          },
        ],
      };
      final listing = Listing.fromJson(json);
      expect(listing.id, 1);
      expect(listing.userId, 42);
      expect(listing.title, 'Fresh honey');
      expect(listing.description, '5kg jars');
      expect(listing.category, 'HONEY');
      expect(listing.price, 12.5);
      expect(listing.quantity, '10 jars');
      expect(listing.address, 'Warsaw');
      expect(listing.lat, 52.2297);
      expect(listing.lng, 21.0122);
      expect(listing.apiaryId, 7);
      expect(listing.apiaryName, 'Home Apiary');
      expect(listing.contactPhone, '123456789');
      expect(listing.contactEmail, 'seller@example.com');
      expect(listing.isHidden, false);
      expect(listing.createdAt, DateTime.parse('2024-02-01T08:30:00Z').toLocal());
      expect(listing.updatedAt, DateTime.parse('2024-02-02T09:00:00Z').toLocal());
      expect(listing.images.length, 1);
      expect(listing.images.first.id, 1);
      expect(listing.distanceKm, 3.25);
    });

    test('handles null price, apiary, and empty images', () {
      final json = {
        'id': 2,
        'user_id': 8,
        'title': 'Beeswax',
        'description': null,
        'category': 'BEESWAX',
        'price': null,
        'quantity': '',
        'address': '',
        'apiary_id': null,
        'apiary_name': null,
        'contact_phone': '',
        'contact_email': '',
        'is_hidden': true,
        'created_at': '2024-03-10T12:00:00Z',
        'updated_at': '2024-03-10T12:00:00Z',
        'images': null,
      };
      final listing = Listing.fromJson(json);
      expect(listing.price, isNull);
      expect(listing.apiaryId, isNull);
      expect(listing.apiaryName, isNull);
      expect(listing.description, '');
      expect(listing.isHidden, true);
      expect(listing.images, isEmpty);
      expect(listing.lat, 0);
      expect(listing.lng, 0);
      expect(listing.distanceKm, isNull);
    });

    test('parses integer price as double', () {
      final json = {
        'id': 3,
        'user_id': 9,
        'title': 'Pollen',
        'description': 'dried pollen',
        'category': 'POLLEN',
        'price': 20,
        'quantity': '1 bag',
        'address': 'Krakow',
        'contact_phone': '987654321',
        'contact_email': 'pollen@example.com',
        'is_hidden': false,
        'created_at': '2024-04-01T10:00:00Z',
        'updated_at': '2024-04-01T10:00:00Z',
      };
      final listing = Listing.fromJson(json);
      expect(listing.price, 20.0);
      expect(listing.price, isA<double>());
    });
  });
}
