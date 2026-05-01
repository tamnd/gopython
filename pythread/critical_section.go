package pythread

import "unsafe"

type criticalFrame struct {
	section  *CriticalSection
	section2 *CriticalSection2
	prev     *criticalFrame
	inactive bool
}

type CriticalSection struct {
	mutex *Mutex
	frame criticalFrame
}

type CriticalSection2 struct {
	base   CriticalSection
	mutex2 *Mutex
}

func (t *ThreadState) BeginCriticalSection(c *CriticalSection, m *Mutex) {
	if m.LockFast() {
		c.mutex = m
		c.frame = criticalFrame{section: c, prev: t.criticalSection}
		t.criticalSection = &c.frame
		return
	}
	t.BeginCriticalSectionSlow(c, m)
}

func (t *ThreadState) BeginCriticalSectionSlow(c *CriticalSection, m *Mutex) {
	if t.criticalSection != nil {
		prev := t.criticalSection
		if prev.section != nil && prev.section.mutex == m {
			c.mutex = nil
			c.frame = criticalFrame{}
			return
		}
		if prev.section2 != nil && prev.section2.mutex2 == m {
			c.mutex = nil
			c.frame = criticalFrame{}
			return
		}
	}
	if t.Interp != nil && t.Interp.StopTheWorld.WorldStopped {
		c.mutex = nil
		c.frame = criticalFrame{}
		return
	}

	c.mutex = nil
	c.frame = criticalFrame{section: c, prev: t.criticalSection}
	t.criticalSection = &c.frame
	m.Lock()
	c.mutex = m
}

func (t *ThreadState) EndCriticalSection(c *CriticalSection) {
	if c.mutex == nil {
		return
	}
	c.mutex.Unlock()
	t.popCriticalSection(c.frame.prev)
}

func (t *ThreadState) BeginCriticalSection2(c *CriticalSection2, m1, m2 *Mutex) {
	if m1 == m2 {
		c.mutex2 = nil
		t.BeginCriticalSection(&c.base, m1)
		return
	}
	if uintptr(unsafe.Pointer(m2)) < uintptr(unsafe.Pointer(m1)) {
		m1, m2 = m2, m1
	}
	if m1.LockFast() {
		if m2.LockFast() {
			c.base.mutex = m1
			c.mutex2 = m2
			c.base.frame = criticalFrame{
				section:  &c.base,
				section2: c,
				prev:     t.criticalSection,
			}
			t.criticalSection = &c.base.frame
			return
		}
		t.BeginCriticalSection2Slow(c, m1, m2, true)
		return
	}
	t.BeginCriticalSection2Slow(c, m1, m2, false)
}

func (t *ThreadState) BeginCriticalSection2Slow(c *CriticalSection2, m1, m2 *Mutex, isM1Locked bool) {
	if t.Interp != nil && t.Interp.StopTheWorld.WorldStopped {
		c.base.mutex = nil
		c.mutex2 = nil
		c.base.frame = criticalFrame{}
		return
	}

	c.base.mutex = nil
	c.mutex2 = nil
	c.base.frame = criticalFrame{
		section:  &c.base,
		section2: c,
		prev:     t.criticalSection,
	}
	t.criticalSection = &c.base.frame

	if !isM1Locked {
		m1.Lock()
	}
	m2.Lock()
	c.base.mutex = m1
	c.mutex2 = m2
}

func (t *ThreadState) EndCriticalSection2(c *CriticalSection2) {
	if c.base.mutex == nil {
		return
	}
	if c.mutex2 != nil {
		c.mutex2.Unlock()
	}
	c.base.mutex.Unlock()
	t.popCriticalSection(c.base.frame.prev)
}

func (t *ThreadState) SuspendCriticalSections() {
	for frame := t.criticalSection; frame != nil; frame = frame.prev {
		if frame.section != nil && frame.section.mutex != nil {
			frame.section.mutex.Unlock()
			if frame.section2 != nil && frame.section2.mutex2 != nil {
				frame.section2.mutex2.Unlock()
			}
		}
		frame.inactive = true
	}
}

func (t *ThreadState) ResumeCriticalSection() {
	frame := t.criticalSection
	if frame == nil || !frame.inactive {
		panic("critical section resume: top frame is active")
	}

	m1 := frame.section.mutex
	frame.section.mutex = nil

	var m2 *Mutex
	if frame.section2 != nil {
		m2 = frame.section2.mutex2
		frame.section2.mutex2 = nil
	}

	if m1 != nil {
		m1.Lock()
	}
	if m2 != nil {
		m2.Lock()
	}

	frame.section.mutex = m1
	if frame.section2 != nil {
		frame.section2.mutex2 = m2
	}
	frame.inactive = false
}

func (t *ThreadState) popCriticalSection(prev *criticalFrame) {
	t.criticalSection = prev
	if prev != nil && prev.inactive {
		t.ResumeCriticalSection()
	}
}
