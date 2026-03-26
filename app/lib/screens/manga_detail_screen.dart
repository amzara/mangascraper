import 'package:flutter/material.dart';
import '../models/manga.dart';
import '../models/manga_detail.dart';
import '../services/api_service.dart';
import 'manga_reader_screen.dart';

class MangaDetailScreen extends StatefulWidget {
  final Manga manga;

  const MangaDetailScreen({super.key, required this.manga});

  @override
  State<MangaDetailScreen> createState() => _MangaDetailScreenState();
}

class _MangaDetailScreenState extends State<MangaDetailScreen> {
  final ApiService _apiService = ApiService();
  MangaDetail? _mangaDetail;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadMangaDetail();
  }

  Future<void> _loadMangaDetail() async {
    try {
      setState(() {
        _isLoading = true;
        _error = null;
      });

      final detail = await _apiService.getMangaDetail(widget.manga.slug);

      setState(() {
        _mangaDetail = detail;
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
        title: Text(widget.manga.title),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadMangaDetail,
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
              onPressed: _loadMangaDetail,
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    if (_mangaDetail == null) {
      return const Center(child: Text('No data'));
    }

    final detail = _mangaDetail!;
    final chapters = detail.chapters;

    return Column(
      children: [
        // Manga info header
        Container(
          padding: const EdgeInsets.all(16),
          color: Theme.of(context).colorScheme.primaryContainer,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                detail.manga.title,
                style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
              ),
              const SizedBox(height: 8),
              Text('Slug: ${detail.manga.slug}'),
              const SizedBox(height: 8),
              Row(
                children: [
                  const Icon(Icons.menu_book, size: 16),
                  const SizedBox(width: 4),
                  Text(
                    '${detail.downloadedCount}/${detail.totalCount} chapters downloaded',
                    style: const TextStyle(fontWeight: FontWeight.w500),
                  ),
                ],
              ),
              const SizedBox(height: 4),
              LinearProgressIndicator(
                value: detail.totalCount > 0
                    ? detail.downloadedCount / detail.totalCount
                    : 0,
                backgroundColor: Colors.grey[300],
                valueColor: AlwaysStoppedAnimation<Color>(
                  detail.downloadedCount == detail.totalCount
                      ? Colors.green
                      : Colors.blue,
                ),
              ),
            ],
          ),
        ),

        // Chapters list
        Expanded(
          child: chapters.isEmpty
              ? const Center(
                  child: Text('No chapters found'),
                )
              : ListView.builder(
                  itemCount: chapters.length,
                  itemBuilder: (context, index) {
                    final chapter = chapters[index];
                    return Card(
                      margin: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 4,
                      ),
                      child: ListTile(
                        leading: Container(
                          width: 12,
                          height: 12,
                          decoration: BoxDecoration(
                            color: Color(chapter.statusColor),
                            shape: BoxShape.circle,
                          ),
                        ),
                        title: Text(chapter.displayName),
                        subtitle: Text(
                          'Status: ${chapter.status}',
                          style: TextStyle(
                            color: Color(chapter.statusColor),
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        trailing: _getStatusIcon(chapter.status),
                        onTap: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) => MangaReaderScreen(
                                chapter: chapter,
                                mangaTitle: detail.manga.title,
                              ),
                            ),
                          );
                        },
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }

  Widget _getStatusIcon(String status) {
    switch (status.toLowerCase()) {
      case 'downloaded':
        return const Icon(Icons.check_circle, color: Colors.green);
      case 'downloading':
        return const Icon(Icons.downloading, color: Colors.blue);
      case 'error':
        return const Icon(Icons.error, color: Colors.red);
      case 'pending':
      default:
        return const Icon(Icons.hourglass_empty, color: Colors.orange);
    }
  }
}
