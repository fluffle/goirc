package state

import (
	"testing"
)

// There is some awkwardness in these tests. Items retrieved directly from the
// state trackers internal maps are private and only have private,
// uncaptialised members. Items retrieved from state tracker public interface
// methods are public and only have public, capitalised members. Comparisons of
// the two are done on the basis of nick or channel name.

func TestSTNewTracker(t *testing.T) {
	st := NewTracker("mynick")

	if len(st.nicks) != 1 {
		t.Errorf("Nick list of new tracker is not 1 (me!).")
	}
	if len(st.chans) != 0 {
		t.Errorf("Channel list of new tracker is not empty.")
	}
	if nk, ok := st.nicks["mynick"]; !ok || nk.nick != "mynick" || nk != st.me {
		t.Errorf("My nick not stored correctly in tracker.")
	}
}

func TestSTNewNick(t *testing.T) {
	st := NewTracker("mynick")
	test1 := st.NewNick("test1")

	if test1 == nil || test1.Nick != "test1" {
		t.Errorf("Nick object created incorrectly by NewNick.")
	}
	if n, ok := st.nicks["test1"]; !ok || !test1.Equals(n.Nick()) || len(st.nicks) != 2 {
		t.Errorf("Nick object stored incorrectly by NewNick.")
	}

	if fail := st.NewNick("test1"); fail != nil {
		t.Errorf("Creating duplicate nick did not produce nil return.")
	}
	if fail := st.NewNick(""); fail != nil {
		t.Errorf("Creating empty nick did not produce nil return.")
	}
}

func TestSTGetNick(t *testing.T) {
	st := NewTracker("mynick")
	test1 := st.NewNick("test1")

	if n := st.GetNick("test1"); !test1.Equals(n) {
		t.Errorf("Incorrect nick returned by GetNick.")
	}
	if n := st.GetNick("test2"); n != nil {
		t.Errorf("Nick unexpectedly returned by GetNick.")
	}
	if len(st.nicks) != 2 {
		t.Errorf("Nick list changed size during GetNick.")
	}
}

func TestSTReNick(t *testing.T) {
	st := NewTracker("mynick")
	test1 := st.NewNick("test1")

	// This channel is here to ensure that its lookup map gets updated
	st.NewChannel("#chan1")
	st.Associate("#chan1", "test1")

	// We need to check out the manipulation of the internals.
	n1 := st.nicks["test1"]
	c1 := st.chans["#chan1"]

	test2 := st.ReNick("test1", "test2")

	if _, ok := st.nicks["test1"]; ok {
		t.Errorf("Nick test1 still exists after ReNick.")
	}
	if n, ok := st.nicks["test2"]; !ok || n != n1 {
		t.Errorf("Nick test2 doesn't exist after ReNick.")
	}
	if _, ok := c1.lookup["test1"]; ok {
		t.Errorf("Channel #chan1 still knows about test1 after ReNick.")
	}
	if n, ok := c1.lookup["test2"]; !ok || n != n1 {
		t.Errorf("Channel #chan1 doesn't know about test2 after ReNick.")
	}
	if test1.Nick != "test1" {
		t.Errorf("Nick test1 changed unexpectedly.")
	}
	if !test2.Equals(n1.Nick()) {
		t.Errorf("Nick test2 did not change.")
	}
	if len(st.nicks) != 2 {
		t.Errorf("Nick list changed size during ReNick.")
	}
	if len(c1.lookup) != 1 {
		t.Errorf("Channel lookup list changed size during ReNick.")
	}

	st.NewNick("test1")
	n2 := st.nicks["test1"]
	fail := st.ReNick("test1", "test2")

	if n, ok := st.nicks["test2"]; !ok || n != n1 {
		t.Errorf("Nick test2 overwritten/deleted by ReNick.")
	}
	if n, ok := st.nicks["test1"]; !ok || n != n2 {
		t.Errorf("Nick test1 overwritten/deleted by ReNick.")
	}
	if fail != nil {
		t.Errorf("ReNick returned Nick on failure.")
	}
	if len(st.nicks) != 3 {
		t.Errorf("Nick list changed size during ReNick.")
	}
}

func TestSTDelNick(t *testing.T) {
	st := NewTracker("mynick")

	add := st.NewNick("test1")
	del := st.DelNick("test1")

	if _, ok := st.nicks["test1"]; ok {
		t.Errorf("Nick test1 still exists after DelNick.")
	}
	if len(st.nicks) != 1 {
		t.Errorf("Nick list still contains nicks after DelNick.")
	}
	if !add.Equals(del) {
		t.Errorf("DelNick returned different nick.")
	}

	// Deleting unknown nick shouldn't work, but let's make sure we have a
	// known nick first to catch any possible accidental removals.
	st.NewNick("test1")
	fail := st.DelNick("test2")
	if fail != nil || len(st.nicks) != 2 {
		t.Errorf("Deleting unknown nick had unexpected side-effects.")
	}

	// Deleting my nick shouldn't work
	fail = st.DelNick("mynick")
	if fail != nil || len(st.nicks) != 2 {
		t.Errorf("Deleting myself had unexpected side-effects.")
	}

	// Test that deletion correctly dissociates nick from channels.
	// NOTE: the two error states in delNick (as opposed to DelNick)
	// are not tested for here, as they will only arise from programming
	// errors in other methods.

	// Create a new channel for testing purposes.
	st.NewChannel("#test1")

	// Associate both "my" nick and test1 with the channel
	st.Associate("#test1", "mynick")
	st.Associate("#test1", "test1")

	// We need to check out the manipulation of the internals.
	n1 := st.nicks["test1"]
	c1 := st.chans["#test1"]

	// Test we have the expected starting state (at least vaguely)
	if len(c1.nicks) != 2 || len(st.nicks) != 2 ||
		len(st.me.chans) != 1 || len(n1.chans) != 1 || len(st.chans) != 1 {
		t.Errorf("Bad initial state for test DelNick() channel dissociation.")
	}

	// Actual deletion tested above...
	st.DelNick("test1")

	if len(c1.nicks) != 1 || len(st.nicks) != 1 ||
		len(st.me.chans) != 1 || len(n1.chans) != 0 || len(st.chans) != 1 {
		t.Errorf("Deleting nick didn't dissociate correctly from channels.")
	}

	if _, ok := c1.nicks[n1]; ok {
		t.Errorf("Nick not removed from channel's nick map.")
	}
	if _, ok := c1.lookup["test1"]; ok {
		t.Errorf("Nick not removed from channel's lookup map.")
	}
}

func TestSTNickInfo(t *testing.T) {
	st := NewTracker("mynick")
	test1 := st.NewNick("test1")
	test2 := st.NickInfo("test1", "foo", "bar", "baz")
	test3 := st.GetNick("test1")

	if test1.Equals(test2) {
		t.Errorf("NickInfo did not return modified nick.")
	}
	if !test3.Equals(test2) {
		t.Errorf("Getting nick after NickInfo returned different nick.")
	}
	test1.Ident, test1.Host, test1.Name = "foo", "bar", "baz"
	if !test1.Equals(test2) {
		t.Errorf("NickInfo did not set nick info correctly.")
	}

	if fail := st.NickInfo("test2", "foo", "bar", "baz"); fail != nil {
		t.Errorf("NickInfo for nonexistent nick did not return nil.")
	}
}

func TestSTNickModes(t *testing.T) {
	st := NewTracker("mynick")
	test1 := st.NewNick("test1")
	test2 := st.NickModes("test1", "+iB")
	test3 := st.GetNick("test1")

	if test1.Equals(test2) {
		t.Errorf("NickModes did not return modified nick.")
	}
	if !test3.Equals(test2) {
		t.Errorf("Getting nick after NickModes returned different nick.")
	}
	test1.Modes.Invisible, test1.Modes.Bot = true, true
	if !test1.Equals(test2) {
		t.Errorf("NickModes did not set nick modes correctly.")
	}

	if fail := st.NickModes("test2", "whatevs"); fail != nil {
		t.Errorf("NickModes for nonexistent nick did not return nil.")
	}
}

func TestSTNewChannel(t *testing.T) {
	st := NewTracker("mynick")

	if len(st.chans) != 0 {
		t.Errorf("Channel list of new tracker is non-zero length.")
	}

	test1 := st.NewChannel("#test1")

	if test1 == nil || test1.Name != "#test1" {
		t.Errorf("Channel object created incorrectly by NewChannel.")
	}
	if c, ok := st.chans["#test1"]; !ok || !test1.Equals(c.Channel()) || len(st.chans) != 1 {
		t.Errorf("Channel object stored incorrectly by NewChannel.")
	}

	if fail := st.NewChannel("#test1"); fail != nil {
		t.Errorf("Creating duplicate chan did not produce nil return.")
	}
	if fail := st.NewChannel(""); fail != nil {
		t.Errorf("Creating empty chan did not produce nil return.")
	}
}

func TestSTGetChannel(t *testing.T) {
	st := NewTracker("mynick")

	test1 := st.NewChannel("#test1")

	if c := st.GetChannel("#test1"); !test1.Equals(c) {
		t.Errorf("Incorrect Channel returned by GetChannel.")
	}
	if c := st.GetChannel("#test2"); c != nil {
		t.Errorf("Channel unexpectedly returned by GetChannel.")
	}
	if len(st.chans) != 1 {
		t.Errorf("Channel list changed size during GetChannel.")
	}
}

func TestSTDelChannel(t *testing.T) {
	st := NewTracker("mynick")

	add := st.NewChannel("#test1")
	del := st.DelChannel("#test1")

	if _, ok := st.chans["#test1"]; ok {
		t.Errorf("Channel test1 still exists after DelChannel.")
	}
	if len(st.chans) != 0 {
		t.Errorf("Channel list still contains chans after DelChannel.")
	}
	if !add.Equals(del) {
		t.Errorf("DelChannel returned different channel.")
	}

	// Deleting unknown channel shouldn't work, but let's make sure we have a
	// known channel first to catch any possible accidental removals.
	st.NewChannel("#test1")
	fail := st.DelChannel("#test2")
	if fail != nil || len(st.chans) != 1 {
		t.Errorf("DelChannel had unexpected side-effects.")
	}

	// Test that deletion correctly dissociates channel from tracked nicks.
	// In order to test this thoroughly we need two channels (so that delNick()
	// is not called internally in delChannel() when len(nick1.chans) == 0.
	st.NewChannel("#test2")
	st.NewNick("test1")

	// Associate both "my" nick and test1 with the channels
	st.Associate("#test1", "mynick")
	st.Associate("#test1", "test1")
	st.Associate("#test2", "mynick")
	st.Associate("#test2", "test1")

	// We need to check out the manipulation of the internals.
	n1 := st.nicks["test1"]
	c1 := st.chans["#test1"]
	c2 := st.chans["#test2"]

	// Test we have the expected starting state (at least vaguely)
	if len(c1.nicks) != 2 || len(c2.nicks) != 2 || len(st.nicks) != 2 ||
		len(st.me.chans) != 2 || len(n1.chans) != 2 || len(st.chans) != 2 {
		t.Errorf("Bad initial state for test DelChannel() nick dissociation.")
	}

	st.DelChannel("#test1")

	// Test intermediate state. We're still on #test2 with test1, so test1
	// shouldn't be deleted from state tracking itself just yet.
	if len(c1.nicks) != 0 || len(c2.nicks) != 2 || len(st.nicks) != 2 ||
		len(st.me.chans) != 1 || len(n1.chans) != 1 || len(st.chans) != 1 {
		t.Errorf("Deleting channel didn't dissociate correctly from nicks.")
	}
	if _, ok := n1.chans[c1]; ok {
		t.Errorf("Channel not removed from nick's chans map.")
	}
	if _, ok := n1.lookup["#test1"]; ok {
		t.Errorf("Channel not removed from nick's lookup map.")
	}

	st.DelChannel("#test2")

	// Test final state. Deleting #test2 means that we're no longer on any
	// common channels with test1, and thus it should be removed from tracking.
	if len(c1.nicks) != 0 || len(c2.nicks) != 0 || len(st.nicks) != 1 ||
		len(st.me.chans) != 0 || len(n1.chans) != 0 || len(st.chans) != 0 {
		t.Errorf("Deleting last channel didn't dissociate correctly from nicks.")
	}
	if _, ok := st.nicks["test1"]; ok {
		t.Errorf("Nick not deleted correctly when on no channels.")
	}
	if _, ok := st.nicks["mynick"]; !ok {
		t.Errorf("My nick deleted incorrectly when on no channels.")
	}
}

func TestSTTopic(t *testing.T) {
	st := NewTracker("mynick")
	test1 := st.NewChannel("#test1")
	test2 := st.Topic("#test1", "foo bar")
	test3 := st.GetChannel("#test1")

	if test1.Equals(test2) {
		t.Errorf("Topic did not return modified channel.")
	}
	if !test3.Equals(test2) {
		t.Errorf("Getting channel after Topic returned different channel.")
	}
	test1.Topic = "foo bar"
	if !test1.Equals(test2) {
		t.Errorf("Topic did not set channel topic correctly.")
	}

	if fail := st.Topic("#test2", "foo baz"); fail != nil {
		t.Errorf("Topic for nonexistent channel did not return nil.")
	}
}

func TestSTChannelModes(t *testing.T) {
	st := NewTracker("mynick")
	test1 := st.NewChannel("#test1")
	test2 := st.ChannelModes("#test1", "+sk", "foo")
	test3 := st.GetChannel("#test1")

	if test1.Equals(test2) {
		t.Errorf("ChannelModes did not return modified channel.")
	}
	if !test3.Equals(test2) {
		t.Errorf("Getting channel after ChannelModes returned different channel.")
	}
	test1.Modes.Secret, test1.Modes.Key = true, "foo"
	if !test1.Equals(test2) {
		t.Errorf("ChannelModes did not set channel modes correctly.")
	}

	if fail := st.ChannelModes("test2", "whatevs"); fail != nil {
		t.Errorf("ChannelModes for nonexistent channel did not return nil.")
	}
}

func TestSTIsOn(t *testing.T) {
	st := NewTracker("mynick")

	st.NewNick("test1")
	st.NewChannel("#test1")

	if priv, ok := st.IsOn("#test1", "test1"); ok || priv != nil {
		t.Errorf("test1 is not on #test1 (yet)")
	}
	st.Associate("#test1", "test1")
	if priv, ok := st.IsOn("#test1", "test1"); !ok || priv == nil {
		t.Errorf("test1 is on #test1 (now)")
	}
}

func TestSTAssociate(t *testing.T) {
	st := NewTracker("mynick")

	st.NewNick("test1")
	st.NewChannel("#test1")

	// We need to check out the manipulation of the internals.
	n1 := st.nicks["test1"]
	c1 := st.chans["#test1"]

	st.Associate("#test1", "test1")
	npriv, nok := n1.chans[c1]
	cpriv, cok := c1.nicks[n1]
	if !nok || !cok || npriv != cpriv {
		t.Errorf("#test1 was not associated with test1.")
	}

	// Test error cases
	if st.Associate("", "test1") != nil {
		t.Errorf("Associating unknown channel did not return nil.")
	}
	if st.Associate("#test1", "") != nil {
		t.Errorf("Associating unknown nick did not return nil.")
	}
	if st.Associate("#test1", "test1") != nil {
		t.Errorf("Associating already-associated things did not return nil.")
	}
}

func TestSTDissociate(t *testing.T) {
	st := NewTracker("mynick")

	st.NewNick("test1")
	st.NewChannel("#test1")
	st.NewChannel("#test2")

	// Associate both "my" nick and test1 with the channels
	st.Associate("#test1", "mynick")
	st.Associate("#test1", "test1")
	st.Associate("#test2", "mynick")
	st.Associate("#test2", "test1")

	// We need to check out the manipulation of the internals.
	n1 := st.nicks["test1"]
	c1 := st.chans["#test1"]
	c2 := st.chans["#test2"]

	// Check the initial state looks mostly like we expect it to.
	if len(c1.nicks) != 2 || len(c2.nicks) != 2 || len(st.nicks) != 2 ||
		len(st.me.chans) != 2 || len(n1.chans) != 2 || len(st.chans) != 2 {
		t.Errorf("Initial state for dissociation tests looks odd.")
	}

	// First, test the case of me leaving #test2
	st.Dissociate("#test2", "mynick")

	// This should have resulted in the complete deletion of the channel.
	if len(c1.nicks) != 2 || len(c2.nicks) != 0 || len(st.nicks) != 2 ||
		len(st.me.chans) != 1 || len(n1.chans) != 1 || len(st.chans) != 1 {
		t.Errorf("Dissociating myself from channel didn't delete it correctly.")
	}
	if st.GetChannel("#test2") != nil {
		t.Errorf("Able to get channel after dissociating myself.")
	}

	// Reassociating myself and test1 to #test2 shouldn't cause any errors.
	st.NewChannel("#test2")
	st.Associate("#test2", "mynick")
	st.Associate("#test2", "test1")

	// c2 is out of date with the complete deletion of the channel
	c2 = st.chans["#test2"]

	// Check state once moar.
	if len(c1.nicks) != 2 || len(c2.nicks) != 2 || len(st.nicks) != 2 ||
		len(st.me.chans) != 2 || len(n1.chans) != 2 || len(st.chans) != 2 {
		t.Errorf("Reassociating to channel has produced unexpected state.")
	}

	// Now, lets dissociate test1 from #test1 then #test2.
	// This first one should only result in a change in associations.
	st.Dissociate("#test1", "test1")

	if len(c1.nicks) != 1 || len(c2.nicks) != 2 || len(st.nicks) != 2 ||
		len(st.me.chans) != 2 || len(n1.chans) != 1 || len(st.chans) != 2 {
		t.Errorf("Dissociating a nick from one channel went wrong.")
	}

	// This second one should also delete test1
	// as it's no longer on any common channels with us
	st.Dissociate("#test2", "test1")

	if len(c1.nicks) != 1 || len(c2.nicks) != 1 || len(st.nicks) != 1 ||
		len(st.me.chans) != 2 || len(n1.chans) != 0 || len(st.chans) != 2 {
		t.Errorf("Dissociating a nick from it's last channel went wrong.")
	}
	if st.GetNick("test1") != nil {
		t.Errorf("Able to get nick after dissociating from all channels.")
	}
}

func TestSTWipe(t *testing.T) {
	st := NewTracker("mynick")

	st.NewNick("test1")
	st.NewNick("test2")
	st.NewNick("test3")
	st.NewChannel("#test1")
	st.NewChannel("#test2")
	st.NewChannel("#test3")

	// Some associations
	st.Associate("#test1", "mynick")
	st.Associate("#test2", "mynick")
	st.Associate("#test3", "mynick")

	st.Associate("#test1", "test1")
	st.Associate("#test2", "test2")
	st.Associate("#test3", "test3")

	st.Associate("#test1", "test2")
	st.Associate("#test2", "test3")

	st.Associate("#test1", "test3")

	// We need to check out the manipulation of the internals.
	nick1 := st.nicks["test1"]
	nick2 := st.nicks["test2"]
	nick3 := st.nicks["test3"]
	chan1 := st.chans["#test1"]
	chan2 := st.chans["#test2"]
	chan3 := st.chans["#test3"]

	// Check the state we have at this point is what we would expect.
	if len(st.nicks) != 4 || len(st.chans) != 3 || len(st.me.chans) != 3 {
		t.Errorf("Tracker nick/channel lists wrong length before wipe.")
	}
	if len(chan1.nicks) != 4 || len(chan2.nicks) != 3 || len(chan3.nicks) != 2 {
		t.Errorf("Channel nick lists wrong length before wipe.")
	}
	if len(nick1.chans) != 1 || len(nick2.chans) != 2 || len(nick3.chans) != 3 {
		t.Errorf("Nick chan lists wrong length before wipe.")
	}

	// Nuke *all* the state!
	st.Wipe()

	// Check the state we have at this point is what we would expect.
	if len(st.nicks) != 1 || len(st.chans) != 0 || len(st.me.chans) != 0 {
		t.Errorf("Tracker nick/channel lists wrong length after wipe.")
	}
	if len(chan1.nicks) != 0 || len(chan2.nicks) != 0 || len(chan3.nicks) != 0 {
		t.Errorf("Channel nick lists wrong length after wipe.")
	}
	if len(nick1.chans) != 0 || len(nick2.chans) != 0 || len(nick3.chans) != 0 {
		t.Errorf("Nick chan lists wrong length after wipe.")
	}
}
