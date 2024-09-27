package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/gordonklaus/portaudio"
)

func main() {
	// 初始化PortAudio
	portaudio.Initialize()
	defer portaudio.Terminate()

	// 设置输入参数
	sampleRate := 44100
	bufferSize := 2048 // 更大的缓冲区
	input := make([]int16, bufferSize)

	// 打开音频流
	stream, err := portaudio.OpenDefaultStream(1, 0, float64(sampleRate), len(input), &input)
	if err != nil {
		fmt.Println("Error opening stream:", err)
		return
	}
	defer stream.Close()

	// 创建WAV文件
	outFile, err := os.Create("output.wav")
	if err != nil {
		fmt.Println("Error creating WAV file:", err)
		return
	}
	defer outFile.Close()

	// 创建WAV编码器
	encoder := wav.NewEncoder(outFile, sampleRate, 16, 1, 1)

	// 启动音频流
	if err := stream.Start(); err != nil {
		fmt.Println("Error starting stream:", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// 使用通道来处理退出
	stopChan := make(chan struct{})

	go func() {
		defer wg.Done()
		fmt.Println("Recording... Press Enter to stop.")

		// 循环读取音频数据
		for {
			select {
			case <-stopChan:
				return
			default:
				if err := stream.Read(); err != nil {
					fmt.Println("Error reading stream:", err)
					return
				}

				// 将int16数据转换为int
				intData := make([]int, len(input))
				for i, v := range input {
					intData[i] = int(v)
				}

				// 使用audio.IntBuffer写入数据
				intBuffer := &audio.IntBuffer{Data: intData, Format: &audio.Format{SampleRate: sampleRate, NumChannels: 1}}

				// 写入音频样本
				if err := encoder.Write(intBuffer); err != nil {
					fmt.Println("Error writing to WAV file:", err)
					return
				}

				// 调整读取速度
				time.Sleep(20 * time.Millisecond) // 控制读取速率
			}
		}
	}()

	// 等待用户输入
	fmt.Scanln()
	close(stopChan) // 发送信号以停止录音

	// 停止音频流
	if err := stream.Stop(); err != nil {
		fmt.Println("Error stopping stream:", err)
	}

	// 等待录音协程完成
	wg.Wait()

	// 关闭编码器
	if err := encoder.Close(); err != nil {
		fmt.Println("Error closing encoder:", err)
	}

	fmt.Println("Saved as output.wav.")
}
