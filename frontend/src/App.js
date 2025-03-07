import React, { useState, useEffect } from 'react';
import axios from 'axios';
import './App.css';

const API_BASE_URL = 'http://localhost:8080/api';

function App() {
  const [books, setBooks] = useState([]);
  const [currentBook, setCurrentBook] = useState(null);
  const [words, setWords] = useState([]);
  const [apiKeySet, setApiKeySet] = useState(true);

  useEffect(() => {
    fetchBooks();
    fetchWords();
  }, []);

  const fetchBooks = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/books`);
      setBooks(response.data || []);
    } catch (error) {
      console.error('Error fetching books:', error);
      setBooks([]);
    }
  };

  const fetchWords = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/words`);
      setWords(response.data.filter(word => word.status !== 'familiar') || []);
    } catch (error) {
      console.error('Error fetching words:', error);
      setWords([]);
    }
  };

  const handleFileUpload = async (event) => {
    const file = event.target.files[0];
    if (!file) return;
    
    const formData = new FormData();
    formData.append('file', file);

    try {
      await axios.post(`${API_BASE_URL}/upload`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      fetchBooks();
    } catch (error) {
      console.error('Error uploading file:', error);
    }
  };

  const handleBookSelect = async (book) => {
    try {
      const response = await axios.get(`${API_BASE_URL}/books/${book.id}`);
      setCurrentBook(response.data);
    } catch (error) {
      console.error('Error fetching book:', error);
    }
  };

  const handleWordClick = async (word) => {
    try {
      const response = await axios.post(`${API_BASE_URL}/translate`, { word });
      if (response.data.translation === "请先设置有道翻译API密钥") {
        setApiKeySet(false);
      }
      const updatedWords = [...words];
      const existingWordIndex = updatedWords.findIndex(w => w.text === word);
      
      if (existingWordIndex >= 0) {
        updatedWords[existingWordIndex] = {
          ...updatedWords[existingWordIndex],
          translation: response.data.translation
        };
      } else {
        updatedWords.push({
          text: word,
          translation: response.data.translation,
          status: response.data.status || 'unfamiliar'
        });
      }
      
      setWords(updatedWords);
    } catch (error) {
      console.error('Error translating word:', error);
    }
  };

  const handleWordStatusChange = async (word, status) => {
    try {
      await axios.post(`${API_BASE_URL}/words`, { ...word, status });
      if (status === 'familiar') {
        setWords(words.filter(w => w.text !== word.text));
      } else {
        fetchWords();
      }
    } catch (error) {
      console.error('Error updating word status:', error);
    }
  };

  const getWordColor = (status) => {
    switch(status) {
      case 'familiar': return 'green';
      case 'learning': return 'orange';
      case 'unfamiliar': return 'red';
      default: return 'black';
    }
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>在线阅读器</h1>
      </header>
      <main>
        {!apiKeySet && (
          <div className="api-key-warning">
            警告：请在后端设置有道翻译API密钥以启用翻译功能。
          </div>
        )}
        <section className="upload-section">
          <h2>上传书籍</h2>
          <input type="file" onChange={handleFileUpload} accept=".txt,.pdf" />
        </section>
        <section className="books-list">
          <h2>书籍列表</h2>
          {books && books.length > 0 ? (
            <ul>
              {books.map(book => (
                <li key={book.id} onClick={() => handleBookSelect(book)}>
                  {book.title}
                </li>
              ))}
            </ul>
          ) : (
            <p>暂无书籍，请上传一本书开始阅读</p>
          )}
        </section>
        <section className="reading-area">
          <h2>阅读区域</h2>
          {currentBook ? (
            <div>
              <h3>{currentBook.book.title}</h3>
              <div className="book-content">
                {currentBook.content && currentBook.content.split('\n').map((paragraph, index) => (
                  <p key={index}>
                    {paragraph.split(/\s+/).map((word, wordIndex) => (
                      <React.Fragment key={`${index}-${wordIndex}`}>
                        <span
                          onClick={() => handleWordClick(word)}
                          className="clickable-word"
                          style={{color: getWordColor(words.find(w => w.text === word)?.status)}}
                        >
                          {word}
                        </span>
                        {wordIndex < paragraph.split(/\s+/).length - 1 && ' '}
                      </React.Fragment>
                    ))}
                  </p>
                ))}
              </div>
            </div>
          ) : (
            <p>请选择一本书开始阅读</p>
          )}
        </section>
        <section className="word-list">
          <h2>单词列表</h2>
          {words && words.length > 0 ? (
            <ul>
              {words.map(word => (
                <li key={word.text}>
                  <span style={{color: getWordColor(word.status)}}>{word.text}</span>
                  <div className="translation">{word.translation}</div>
                  <select
                    value={word.status || 'unfamiliar'}
                    onChange={(e) => handleWordStatusChange(word, e.target.value)}
                  >
                    <option value="unfamiliar">不熟悉</option>
                    <option value="learning">学习中</option>
                    <option value="familiar">熟悉</option>
                  </select>
                </li>
              ))}
            </ul>
          ) : (
            <p>暂无单词，点击阅读区域中的单词开始学习</p>
          )}
        </section>
      </main>
    </div>
  );
}

export default App;
