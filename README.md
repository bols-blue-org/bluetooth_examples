
# bluetoothのサンプルコード

## 環境
ホスト: linux ubuntu 18.04
BLE デバイス: thingy 52

## ビルド
go build ./paypal/gatt/paypal_thingy/
go build ./currantlabs/ble/currantlabs_thingy/

## 実行
sudo ./paypal_thingy DF:A6:49:DD:81:87
sudo ./currantlabs_thingy -sd 30s -addr DF:A6:49:DD:81:87

## 現在の問題点

BLEのgatt 通信で読み込みはできるが書き込み時に書き込めていない
BLE Scannerは書き込みできているのでデバイス自体の問題ではない模様
