// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package miner

import (
	"sync"

	"sync/atomic"

	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/log"
)

type CpuAgent struct {
	mu sync.Mutex

	workCh        chan *Work //接收挖矿任务管道
	stop          chan struct{} //停止管道
	quitCurrentOp chan struct{} //停止当前挖矿操作的管道
	returnCh      chan<- *Result //worker通过这个管道接受挖完的区块

	chain  consensus.ChainReader
	engine consensus.Engine

	isMining int32 // isMining indicates whether the agent is currently mining
}

//主要功能:根据掺入参数或默认参数实例化CpuAgent类
//task1:根据掺入参数或默认参数实例化CpuAgent类
func NewCpuAgent(chain consensus.ChainReader, engine consensus.Engine) *CpuAgent {
	//--------------------------------------task1--------------------------------------
	//task1:根据掺入参数或默认参数实例化CpuAgent类
	// ---------------------------------------------------------------------------------
	miner := &CpuAgent{
		chain:  chain,
		engine: engine,
		stop:   make(chan struct{}, 1),
		workCh: make(chan *Work, 1),
	}
	return miner
}

func (self *CpuAgent) Work() chan<- *Work            { return self.workCh }
func (self *CpuAgent) SetReturnCh(ch chan<- *Result) { self.returnCh = ch }

func (self *CpuAgent) Stop() {
	if !atomic.CompareAndSwapInt32(&self.isMining, 1, 0) {
		return // agent already stopped
	}
	self.stop <- struct{}{}
done:
	// Empty work channel
	for {
		select {
		case <-self.workCh:
		default:
			break done
		}
	}
}
//主要功能:启动接受事件go程等待挖矿任务
//task1:启动接受事件go程等待挖矿任务
func (self *CpuAgent) Start() {
	if !atomic.CompareAndSwapInt32(&self.isMining, 0, 1) {
		return // agent already started
	}
	//--------------------------------------task1--------------------------------------
	//task1:启动接受事件go程等待挖矿任务
	// ---------------------------------------------------------------------------------
	go self.update()
}

//主要功能:接收挖矿任务事件和挖矿停止事件
//task1:接收挖矿任务事件
//task2:接受停止挖矿事件
func (self *CpuAgent) update() {
out:
	for {
		select {
		//--------------------------------------task1--------------------------------------
		//task1:接收挖矿任务事件
		// ---------------------------------------------------------------------------------
		case work := <-self.workCh:
			self.mu.Lock()
			//如果退出任务管道不为空,就关闭它
			if self.quitCurrentOp != nil {
				close(self.quitCurrentOp)
			}
			self.quitCurrentOp = make(chan struct{})
			//调用一致性引擎进行挖矿
			go self.mine(work, self.quitCurrentOp)
			self.mu.Unlock()
		//--------------------------------------task1--------------------------------------
		//task2:接受停止挖矿事件
		// ---------------------------------------------------------------------------------
		case <-self.stop:
			self.mu.Lock()
			if self.quitCurrentOp != nil {
				close(self.quitCurrentOp)
				self.quitCurrentOp = nil
			}
			self.mu.Unlock()
			break out
		}
	}
}

//主要功能:调用一致性引擎进行挖矿
//task1:如果挖矿成功，则将结果通过returnCh返回给Worker
//task2:如果发生错误，则将nil通过returnCh返回给Worker
func (self *CpuAgent) mine(work *Work, stop <-chan struct{}) {
	if result, err := self.engine.Seal(self.chain, work.Block, stop); result != nil {
		log.Info("Successfully sealed new block", "number", result.Number(), "hash", result.Hash())
		self.returnCh <- &Result{work, result}
	} else {
		if err != nil {
			log.Warn("Block sealing failed", "err", err)
		}
		self.returnCh <- nil
	}
}

func (self *CpuAgent) GetHashRate() int64 {
	if pow, ok := self.engine.(consensus.PoW); ok {
		return int64(pow.Hashrate())
	}
	return 0
}
