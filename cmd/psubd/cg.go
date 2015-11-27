package main

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/funkygao/log4go"
	"github.com/wvanbergen/kafka/consumergroup"
)

type consumerGroups struct {
	hostname  string
	metaStore MetaStore

	shutdownCh chan struct{}

	// {topic: {group: {clientId: consumerGroup}}}
	consumerGroups map[string]map[string]map[string]*consumergroup.ConsumerGroup
	cgLock         sync.RWMutex

	// {topic: {group: {clientId: {partitionId: message}}}}
	consumerOffsets map[string]map[string]map[string]map[int32]*sarama.ConsumerMessage
	offsetsLock     sync.Mutex
}

func newConsumerGroups(hostname string, metaStore MetaStore, shutdownCh chan struct{}) *consumerGroups {
	return &consumerGroups{
		hostname:        hostname,
		metaStore:       metaStore,
		shutdownCh:      shutdownCh,
		consumerGroups:  make(map[string]map[string]map[string]*consumergroup.ConsumerGroup),
		consumerOffsets: make(map[string]map[string]map[string]map[int32]*sarama.ConsumerMessage),
	}
}

// TODO resume from last offset
func (this *consumerGroups) pickConsumerGroup(topic, group,
	clientId string) (cg *consumergroup.ConsumerGroup, err error) {
	this.cgLock.Lock()
	defer this.cgLock.Unlock()

	var present bool
	if _, present = this.consumerGroups[topic]; !present {
		this.consumerGroups[topic] = make(map[string]map[string]*consumergroup.ConsumerGroup)
	}
	if _, present = this.consumerGroups[topic][group]; !present {
		this.consumerGroups[topic][group] = make(map[string]*consumergroup.ConsumerGroup)
	}
	cg, present = this.consumerGroups[topic][group][clientId]
	if present {
		log.Debug("found cg for %s:%s:%s", topic, group, clientId)
		return
	}

	if len(this.consumerGroups[topic][group]) >= len(this.metaStore.Partitions(topic)) {
		err = ErrTooManyConsumers
		log.Error("topic:%s group:%s client:%s %v", topic, group, clientId, err)

		return
	}

	// create the consumer group for this client
	cf := consumergroup.NewConfig()
	cf.Zookeeper.Chroot = this.metaStore.ZkChroot()
	cf.ClientID = this.hostname
	cg, err = consumergroup.JoinConsumerGroup(group, []string{topic},
		this.metaStore.ZkAddrs(), cf)
	if err == nil {
		this.consumerGroups[topic][group][clientId] = cg
	}

	return
}

// TODO
func (this *consumerGroups) trackOffset(topic, group, client string, message *sarama.ConsumerMessage) {
	this.offsetsLock.Lock()
	defer this.offsetsLock.Unlock()

	var present bool
	if _, present = this.consumerOffsets[topic]; !present {
		this.consumerOffsets[topic] = make(map[string]map[string]map[int32]*sarama.ConsumerMessage)
	}
	if _, present = this.consumerOffsets[topic][group]; !present {
		this.consumerOffsets[topic][group] = make(map[string]map[int32]*sarama.ConsumerMessage)
	}
	if _, present = this.consumerOffsets[topic][group][client]; !present {
		this.consumerOffsets[topic][group][client] = make(map[int32]*sarama.ConsumerMessage)
	}

	this.consumerOffsets[message.Topic][group][client][message.Partition] = message
}

func (this *consumerGroups) start() {
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-this.shutdownCh:
			break

		case <-ticker.C:
			// TODO thundering herd
			for topic, ts := range this.consumerOffsets {
				for group, gs := range ts {
					for client, cs := range gs {
						for _, msg := range cs {
							this.consumerGroups[topic][group][client].CommitUpto(msg)
						}
					}
				}
			}

		}
	}

}
