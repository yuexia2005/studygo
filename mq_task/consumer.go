package mq_task

import (
	"encoding/json"
	"log"
	"video_feed/models"

	"gorm.io/gorm"
)

func StartConsumer() {
	//调用信道 Consume 方法，从队列 video_tasks 消费消息
	msgs, err := models.MQChannel.Consume(
		"video_tasks",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("监听队列失败: %v", err)
	}

	go func() {
		for d := range msgs {
			var data map[string]interface{}
			//反序列化存入data
			json.Unmarshal(d.Body, &data)

			//从data取点赞状态片段并断言
			taskType, ok := data["task_type"].(string)
			// 如果不是点赞任务，走原来的视频处理逻辑
			if !ok || taskType == "" {
				log.Printf(" [MQ] 处理视频任务: VideoID=%v", data["video_id"])
				d.Ack(false)
				continue
			}

			//解析ID（JSON解析数字默认为float64）
			uID := uint(data["user_id"].(float64))
			vID := uint(data["video_id"].(float64))

			// 开启事务保证 Likes 表和 Videos 表同步
			err := models.DB.Transaction(func(tx *gorm.DB) error {
				//未点赞增加点赞
				if taskType == "like" {
					if err := tx.FirstOrCreate(&models.Like{}, models.Like{UserID: uID, VideoID: vID}).Error; err != nil {
						return err
					}
					//视频表点赞数+1
					return tx.Model(&models.Video{}).Where("id = ?", vID).Update("like_count", gorm.Expr("like_count + 1")).Error
				} else {
					//点赞了取消点赞
					//删除记录
					if err := tx.Where("user_id = ? AND video_id = ?", uID, vID).Delete(&models.Like{}).Error; err != nil {
						return err
					}
					//视频表里点赞数 -1
					return tx.Model(&models.Video{}).Where("id = ? AND like_count >0", vID).Update("like_count", gorm.Expr("like_count - 1")).Error
				}

			})

			if err != nil {
				log.Printf("[MQ]数据库同步失败：%v", err)
			} else {
				log.Printf("[MQ]成功异步同步点赞状态：用户%d %s 视频%d", uID, taskType, vID)
			}

			d.Ack(false) // 确认消息处理完毕
		}
	}()
}
