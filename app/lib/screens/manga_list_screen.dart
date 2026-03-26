import 'package:flutter/material.dart';
import '../models/manga.dart';
import '../services/api_service.dart';
import 'manga_detail_screen.dart';

class MangaListScreen extends StatefulWidget {
  const MangaListScreen({super.key});

  @override
  State<MangaListScreen> createState() => _MangaListScreenState();
}

class _MangaListScreenState extends State<MangaListScreen> {
  final ApiService _apiService = ApiService();
  List<Manga> _mangaList = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadManga();
  }

  Future<void> _loadManga() async {
    try {
      setState(() {
        _isLoading = true;
        _error = null;
      });

      final manga = await _apiService.getAllManga();

      setState(() {
        _mangaList = manga;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('My Manga Collection'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadManga,
          ),
        ],
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_error != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text('Error: $_error', textAlign: TextAlign.center),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: _loadManga,
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    if (_mangaList.isEmpty) {
      return const Center(
        child: Text('No manga found. Go back and add some!'),
      );
    }

    return ListView.builder(
      itemCount: _mangaList.length,
      itemBuilder: (context, index) {
        final manga = _mangaList[index];
        return Card(
          margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: ListTile(
            leading: CircleAvatar(
              child: Text('${index + 1}'),
            ),
            title: Text(
              manga.title,
              style: const TextStyle(fontWeight: FontWeight.bold),
            ),
            subtitle: Text('Slug: ${manga.slug}'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (context) => MangaDetailScreen(manga: manga),
                ),
              );
            },
          ),
        );
      },
    );
  }
}
