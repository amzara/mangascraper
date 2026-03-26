import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/manga.dart';
import '../models/manga_detail.dart';
import '../models/chapter_pages.dart';
import '../models/search_result.dart';

class ApiService {
  static const String baseUrl = 'http://a:8081';

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
      final response = await http.get(
        Uri.parse('$baseUrl/api/chapter/$chapterId'),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['success'] == true) {
          return ChapterPages.fromJson(data);
        }
      }
      throw Exception('Failed to load chapter pages: ${response.statusCode}');
    } catch (e) {
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
}
