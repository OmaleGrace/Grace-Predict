import requests
from sklearn.linear_model import LogisticRegression
from sklearn.model_selection import train_test_split

url = 'https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=IBM&apikey=ZZSNUWC1CI2R7C0X'
r = requests.get(url)
data = r.json()

time_series = data['Time Series (Daily)']
dates = sorted(time_series.keys())  # oldest to newest

closes, volumes, highs, lows, opens = [], [], [], [], []
for date in dates:
    day = time_series[date]
    closes.append(float(day['4. close']))
    volumes.append(float(day['5. volume']))
    highs.append(float(day['2. high']))
    lows.append(float(day['3. low']))
    opens.append(float(day['1. open']))

X = []
y = []

# start at index 5 so we can compute a 5-day moving average
for i in range(5, len(closes) - 1):
    pct_change = (closes[i] - closes[i-1]) / closes[i-1]
    moving_avg_5 = sum(closes[i-5:i]) / 5
    high_low_spread = highs[i] - lows[i]
    close_open_diff = closes[i] - opens[i]
    volume = volumes[i]

    features = [pct_change, moving_avg_5, high_low_spread, close_open_diff, volume]
    label = 1 if closes[i+1] > closes[i] else 0  # did tomorrow go up?

    X.append(features)
    y.append(label)

X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, shuffle=False)

model = LogisticRegression(max_iter=1000)
model.fit(X_train, y_train)

accuracy = model.score(X_test, y_test)
print("Accuracy:", accuracy)