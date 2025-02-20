package notification

import (
	"fmt"
	"time"

	"github.com/kridavyuha/api-server/pkg/kvstore"
	"gorm.io/gorm"
)

type NotificationService struct {
	KV kvstore.KVStore
	DB *gorm.DB
}

func New(kv kvstore.KVStore, db *gorm.DB) *NotificationService {
	return &NotificationService{
		KV: kv,
		DB: db,
	}
}

type NotificationDetails struct {
	Id           int       `json:"id"`
	ActorId      int       `json:"actor_id"`
	EntityTypeId int       `json:"entity_type_id"`
	EntityId     int       `json:"entity_id"`
	CreatedAt    time.Time `json:"created_at"`
	Status       string    `json:"status"`
}

func (ns *NotificationService) GetNotifications(user_id int) ([]Notification, error) {

	var details []NotificationDetails

	err := ns.DB.Raw(`SELECT N.id, N.actor_id, N.status, NO.entity_type_id, NO.entity_id, NO.created_at from notification as N join notification_obj as NO on N.notification_obj_id = NO.id where N.notifier_id = ?`, user_id).Scan(&details).Error

	if err != nil {
		return nil, err
	}

	// loop over the notificaiton details and if entity_type_id ==1 || == 2 get the txn record.
	notifications := make([]Notification, 0)

	for _, notif := range details {
		notification := Notification{
			Actor:  1,
			Status: notif.Status,
			Id:     notif.Id,
		}
		if notif.EntityTypeId == 1 || notif.EntityTypeId == 2 {
			notification.Entity = "transaction"
			// get the txn:
			var transaction Transaction

			err := ns.DB.Raw("SELECT T.shares, T.price, T.transaction_time, P.player_name from transactions as T join players as P on P.player_id = T.player_id  where T.id = ?", notif.EntityId).Scan(&transaction).Error

			if err != nil {
				return nil, err
			}

			if notif.EntityTypeId == 1 {
				// buy
				notification.Description = fmt.Sprintf("Your order to buy %v at %.2f executed successfully", transaction.PlayerName, transaction.Price)
			} else {
				notification.Description = fmt.Sprintf("Your order to sell %v at %.2f executed successfully", transaction.PlayerName, transaction.Price)

			}
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil

}

func (ns *NotificationService) UpdateNotificationStatus(user_id int) error {
	err := ns.DB.Exec("UPDATE notification SET status = ? where notifier_id = ?", "seen", user_id).Error
	if err != nil {
		return fmt.Errorf("not able to update status of notification with err: %v", err)
	}
	return nil
}
