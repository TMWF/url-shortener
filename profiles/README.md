### Выводы по инкременту 17

Замена использования encoding/json на "github.com/goccy/go-json" и прямая запись в http.ResponseWriter через json.NewEncoder().Encode() аозволило немного увеличить скорость работы хендлеров и уменьшить потребление памяти.

File: handler.test
Type: alloc_space
Time: 2026-06-14 16:50:02 MSK
Showing nodes accounting for -893.64MB, 2.64% of 33802.66MB total
Dropped 75 nodes (cum <= 169.01MB)
      flat  flat%   sum%        cum   cum%
 -783.38MB  2.32%  2.32%  -783.38MB  2.32%  encoding/json.(*Decoder).refill
 -524.66MB  1.55%  3.87%  -524.66MB  1.55%  encoding/json.NewDecoder (inline)
  523.71MB  1.55%  2.32%   523.71MB  1.55%  github.com/goccy/go-json/internal/decoder.NewStream (inline)
  417.70MB  1.24%  1.08%   417.70MB  1.24%  io.ReadAll
 -305.04MB   0.9%  1.99%  -322.05MB  0.95%  encoding/json.Marshal
 -140.56MB  0.42%  2.40%  -140.56MB  0.42%  bufio.NewReaderSize (inline)
 -112.01MB  0.33%  2.73%  -112.01MB  0.33%  encoding/json.(*decodeState).literalStore
 -107.50MB  0.32%  3.05%  -219.51MB  0.65%  encoding/json.(*decodeState).object
  -94.01MB  0.28%  3.33%   -94.01MB  0.28%  reflect.growslice
      90MB  0.27%  3.06%       90MB  0.27%  reflect.unsafe_New
   86.01MB  0.25%  2.81%    86.01MB  0.25%  reflect.unsafe_NewArray
  -77.03MB  0.23%  3.04%   -77.03MB  0.23%  net/textproto.MIMEHeader.Add (inline)
   76.04MB  0.22%  2.81%    76.04MB  0.22%  bytes.growSlice
     -66MB   0.2%  3.01%      -66MB   0.2%  net/http.Header.Clone (inline)
   65.01MB  0.19%  2.82%    72.18MB  0.21%  github.com/goccy/go-json.unmarshal
   64.01MB  0.19%  2.63%    64.01MB  0.19%  github.com/TMWF/url-shortener/internal/handler_test.BenchmarkShortenURLBatch.func1
     -55MB  0.16%  2.79%      -55MB  0.16%  encoding/json.(*scanner).pushParseState
  -51.50MB  0.15%  2.94%   -51.50MB  0.15%  net/http/httptest.NewRecorder (inline)
   37.03MB  0.11%  2.83%    37.03MB  0.11%  github.com/goccy/go-json/internal/encoder.init.func1
  -37.01MB  0.11%  2.94%   -36.51MB  0.11%  net/http.readRequest
      31MB 0.092%  2.85%    78.01MB  0.23%  net/http/httptest.(*ResponseRecorder).Result
      30MB 0.089%  2.76%       30MB 0.089%  bytes.NewReader (inline)
      24MB 0.071%  2.69%       24MB 0.071%  io.LimitReader (inline)
   23.02MB 0.068%  2.62%    23.02MB 0.068%  sync.(*Pool).pinSlow
   21.50MB 0.064%  2.56%    21.50MB 0.064%  strings.NewReader (inline)
  -13.49MB  0.04%  2.60%   -13.49MB  0.04%  net/textproto.MIMEHeader.Set (inline)
     -13MB 0.038%  2.64%      -13MB 0.038%  internal/bytealg.MakeNoZero
     -12MB 0.036%  2.67%      -23MB 0.068%  github.com/TMWF/url-shortener/internal/handler.(*urlHandler).getContextWithUserIDIfNeeded
      12MB 0.036%  2.64%       12MB 0.036%  io.NopCloser (inline)
     -11MB 0.033%  2.67%      -11MB 0.033%  context.WithValue
      -9MB 0.027%  2.70%       -9MB 0.027%  github.com/TMWF/url-shortener/internal/handler_test.BenchmarkGetUserURLs_Parallel.func1
    6.50MB 0.019%  2.68%    11.01MB 0.033%  fmt.Sprintf
    5.50MB 0.016%  2.66%     5.50MB 0.016%  net/url.parse
    5.50MB 0.016%  2.64%  -377.39MB  1.12%  github.com/TMWF/url-shortener/internal/handler.(*urlHandler).ShortenURLAPI
    5.50MB 0.016%  2.63%     5.50MB 0.016%  github.com/TMWF/url-shortener/internal/handler_test.BenchmarkShortenURLAPI.func1
       5MB 0.015%  2.61%        5MB 0.015%  github.com/TMWF/url-shortener/internal/handler_test.BenchmarkShortenURLBatch_Parallel.func1
   -4.50MB 0.013%  2.63%    -4.50MB 0.013%  net/textproto.readMIMEHeader
      -4MB 0.012%  2.64%  -151.54MB  0.45%  github.com/TMWF/url-shortener/internal/handler.(*urlHandler).ShortenURL
   -3.50MB  0.01%  2.65%  -395.56MB  1.17%  github.com/TMWF/url-shortener/internal/handler.(*urlHandler).GetUserURLs
    1.50MB 0.0044%  2.64%  -152.06MB  0.45%  net/http/httptest.NewRequestWithContext
      -1MB 0.003%  2.65%   -14.51MB 0.043%  encoding/json.newEncodeState
    0.50MB 0.0015%  2.65%    76.54MB  0.23%  bytes.(*Buffer).grow
    0.50MB 0.0015%  2.64%  -333.34MB  0.99%  github.com/TMWF/url-shortener/internal/handler_test.BenchmarkShortenURL
         0     0%  2.64%  -140.56MB  0.42%  bufio.NewReader (inline)
         0     0%  2.64%    77.54MB  0.23%  bytes.(*Buffer).Write
         0     0%  2.64% -1151.90MB  3.41%  encoding/json.(*Decoder).Decode
         0     0%  2.64%  -816.88MB  2.42%  encoding/json.(*Decoder).readValue
         0     0%  2.64%  -203.01MB   0.6%  encoding/json.(*decodeState).array
         0     0%  2.64%   -21.50MB 0.064%  encoding/json.(*decodeState).scanWhile
         0     0%  2.64%  -335.02MB  0.99%  encoding/json.(*decodeState).unmarshal
         0     0%  2.64%  -326.52MB  0.97%  encoding/json.(*decodeState).value
         0     0%  2.64%      -55MB  0.16%  encoding/json.stateBeginValue
         0     0%  2.64%   -27.50MB 0.081%  encoding/json.stateBeginValueOrEmpty
         0     0%  2.64%   124.58MB  0.37%  github.com/TMWF/url-shortener/internal/handler.(*urlHandler).ShortenURLBatch
         0     0%  2.64%   -90.03MB  0.27%  github.com/TMWF/url-shortener/internal/handler.(*urlHandler).setUserJWTCookieIfNeeded
         0     0%  2.64%      -11MB 0.033%  github.com/TMWF/url-shortener/internal/handler_test.(*mockURLService).GetUserURLs
         0     0%  2.64%        7MB 0.021%  github.com/TMWF/url-shortener/internal/handler_test.(*mockURLService).ShortenURLAPI
         0     0%  2.64%    69.01MB   0.2%  github.com/TMWF/url-shortener/internal/handler_test.(*mockURLService).ShortenURLBatch
         0     0%  2.64%  -127.06MB  0.38%  github.com/TMWF/url-shortener/internal/handler_test.BenchmarkGetUserURLs
         0     0%  2.64% -1022.69MB  3.03%  github.com/TMWF/url-shortener/internal/handler_test.BenchmarkGetUserURLs_Parallel.func4
         0     0%  2.64%   362.28MB  1.07%  github.com/TMWF/url-shortener/internal/handler_test.BenchmarkShortenURLAPI
         0     0%  2.64%  -553.20MB  1.64%  github.com/TMWF/url-shortener/internal/handler_test.BenchmarkShortenURLAPI_Parallel.func4
         0     0%  2.64%   471.43MB  1.39%  github.com/TMWF/url-shortener/internal/handler_test.BenchmarkShortenURLBatch
         0     0%  2.64%   417.70MB  1.24%  github.com/TMWF/url-shortener/internal/handler_test.BenchmarkShortenURLBatch_Parallel.func4
         0     0%  2.64%  -112.59MB  0.33%  github.com/TMWF/url-shortener/internal/handler_test.BenchmarkShortenURL_Parallel.func4
         0     0%  2.64%   180.52MB  0.53%  github.com/goccy/go-json.(*Decoder).Decode (inline)
         0     0%  2.64%   180.52MB  0.53%  github.com/goccy/go-json.(*Decoder).DecodeWithOption
         0     0%  2.64%   467.63MB  1.38%  github.com/goccy/go-json.(*Encoder).Encode (inline)
         0     0%  2.64%   467.63MB  1.38%  github.com/goccy/go-json.(*Encoder).EncodeWithOption
         0     0%  2.64%   408.58MB  1.21%  github.com/goccy/go-json.(*Encoder).encodeWithOption
         0     0%  2.64%   523.71MB  1.55%  github.com/goccy/go-json.NewDecoder (inline)
         0     0%  2.64%    72.18MB  0.21%  github.com/goccy/go-json.Unmarshal (inline)
         0     0%  2.64%   180.52MB  0.53%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).DecodeStream
         0     0%  2.64%     6.01MB 0.018%  github.com/goccy/go-json/internal/decoder.TakeRuntimeContext (inline)
         0     0%  2.64%    58.05MB  0.17%  github.com/goccy/go-json/internal/encoder.TakeRuntimeContext (inline)
         0     0%  2.64%      -13MB 0.038%  net/http.(*Cookie).String
         0     0%  2.64%   -77.03MB  0.23%  net/http.Header.Add (inline)
         0     0%  2.64%   -13.49MB  0.04%  net/http.Header.Set (inline)
         0     0%  2.64%   -36.51MB  0.11%  net/http.ReadRequest
         0     0%  2.64%   -90.03MB  0.27%  net/http.SetCookie
         0     0%  2.64%    79.04MB  0.23%  net/http/httptest.(*ResponseRecorder).Write
         0     0%  2.64%      -66MB   0.2%  net/http/httptest.(*ResponseRecorder).WriteHeader
         0     0%  2.64%  -152.06MB  0.45%  net/http/httptest.NewRequest (inline)
         0     0%  2.64%    -4.50MB 0.013%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0%  2.64%     5.50MB 0.016%  net/url.ParseRequestURI
         0     0%  2.64%   -94.01MB  0.28%  reflect.Value.Grow
         0     0%  2.64%   -94.01MB  0.28%  reflect.Value.grow
         0     0%  2.64%      -13MB 0.038%  strings.(*Builder).Grow
         0     0%  2.64%      -13MB 0.038%  strings.(*Builder).grow
         0     0%  2.64%    59.55MB  0.18%  sync.(*Pool).Get
         0     0%  2.64%    23.02MB 0.068%  sync.(*Pool).pin
         0     0%  2.64% -1270.79MB  3.76%  testing.(*B).RunParallel.func1
         0     0%  2.64%   372.64MB  1.10%  testing.(*B).launch
         0     0%  2.64%   372.31MB  1.10%  testing.(*B).runN