import 'dart:convert';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:http/http.dart' as http;
import '../models/manga.dart';
import '../models/manga_detail.dart';
import '../models/chapter_pages.dart';
import '../models/search_result.dart';

class ApiService {
  static String get baseUrl {
    final host = dotenv.env['API_HOST']?.trim();
    if (host == null || host.isEmpty) {
      throw Exception(
        'API_HOST is not set. Please copy app/.env.example to app/.env '
        'and set your hostname (e.g. your Tailscale machine name).',
      );
    }
    return 'http://$host:8081';
  }

  Future<List<Manga>> getAllManga() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/api/manga'),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['success'] == true) {
          final List<dynamic> mangaList = data['data'];
          return mangaList.map((json) => Manga.fromJson(json)).toList();
        }
      }
      throw Exception('Failed to load manga: ${response.statusCode}');
    } catch (e) {
      throw Exception('Error fetching manga: $e');
    }
  }

  Future<MangaDetail> getMangaDetail(String slug) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/api/manga/$slug'),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['success'] == true) {
          return MangaDetail.fromJson(data);
        }
      }
      throw Exception('Failed to load manga detail: ${response.statusCode}');
    } catch (e) {
      throw Exception('Error fetching manga detail: $e');
    }
  }

  Future<ChapterPages> getChapterPages(int chapterId) async {
    try {
      final uri = Uri.parse('$baseUrl/api/chapter/$chapterId');
      print('[ApiService] Fetching chapter pages from: $uri');
      final response = await http.get(
        uri,
        headers: {'Content-Type': 'application/json'},
      );

      print('[ApiService] Response status: ${response.statusCode}');
      if (response.statusCode == 200) {
        final body = response.body;
        print('[ApiService] Raw response body: $body');
        final data = jsonDecode(body);
        print('[ApiService] Response data keys: ${data.keys.toList()}');
        if (data['success'] == true) {
          print('[ApiService] chapter json: ${data['chapter']}');
          print('[ApiService] pages count: ${(data['pages'] as List?)?.length}');
          if (data['pages'] != null && (data['pages'] as List).isNotEmpty) {
            print('[ApiService] first page json: ${(data['pages'] as List).first}');
          }
          try {
            return ChapterPages.fromJson(data);
          } catch (parseError) {
            print('[ApiService] JSON parse error: $parseError');
            print('[ApiService] Full data that failed: $data');
            rethrow;
          }
        }
      }
      throw Exception('Failed to load chapter pages: ${response.statusCode}');
    } catch (e, stack) {
      print('[ApiService] Error fetching chapter pages: $e');
      print('[ApiService] Stack trace: $stack');
      throw Exception('Error fetching chapter pages: $e');
    }
  }

  Future<void> getToken() async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/api/token'),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['success'] == true) {
          return;
        }
      }
      throw Exception('Failed to get token: ${response.statusCode}');
    } catch (e) {
      throw Exception('Error getting token: $e');
    }
  }

  Future<List<SearchResult>> searchManga(String query) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/api/search?q=$query'),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['success'] == true) {
          final List<dynamic> results = data['data'];
          return results.map((json) => SearchResult.fromJson(json)).toList();
        }
      }
      throw Exception('Failed to search manga: ${response.statusCode}');
    } catch (e) {
      throw Exception('Error searching manga: $e');
    }
  }

  Future<void> downloadManga(String slug) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/api/manga/download'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'title': slug}),
      );

      if (response.statusCode != 202) {
        throw Exception('Failed to start download: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error starting download: $e');
    }
  }

  Future<void> deleteManga(int mangaId) async {
    try {
      final response = await http.delete(
        Uri.parse('$baseUrl/api/manga/delete/$mangaId'),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode != 200) {
        throw Exception('Failed to delete manga: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error deleting manga: $e');
    }
  }
}
