package content

// inSign holds hand-written texts for each body/point in each sign, in zodiac
// order (Aries … Pisces).
var inSign = map[string][12]string{
	"sun": {
		"With your Sun in Aries, you come alive when you are starting something. You are direct, brave and quick to act, and you would rather try and fail than wait around. You need challenges and a sense of independence. Learning to pause before you leap turns your courage into real leadership.",
		"With your Sun in Taurus, you are steady, patient and grounded. You build your life step by step and enjoy the good things it brings: comfort, beauty, good food and people you can trust. Once you commit, you stay. Your growth comes from welcoming change before it is forced on you.",
		"With your Sun in Gemini, your mind is always on the move. You are curious, chatty and quick to pick up new ideas, and you enjoy connecting people and information. Variety keeps you happy. Learning to stay with one thing a little longer helps your talents go deep.",
		"With your Sun in Cancer, you are caring, intuitive and devoted to the people you love. Home and family, whether given or chosen, matter more than almost anything. You feel things deeply and remember everything. Your strength grows when you look after yourself as well as you look after others.",
		"With your Sun in Leo, you are warm, generous and made to shine. You love to create, to celebrate and to make people feel special. You want to be appreciated, and you give your own praise freely. Your confidence grows strongest when it no longer depends on applause.",
		"With your Sun in Virgo, you are thoughtful, practical and quietly helpful. You notice details and like doing things properly, and you feel best when you are useful. You often see how something could be better. Being as kind to yourself as you are careful with your work is your main lesson.",
		"With your Sun in Libra, you are gracious, fair-minded and made for partnership. You see every side of a question and want everyone to be treated well. Beauty and harmony lift your mood. Your growth comes from trusting your own choices, even when not everyone agrees.",
		"With your Sun in Scorpio, you are intense, private and determined. You care about the truth beneath the surface and you love with total loyalty. You are far more resilient than you look. Learning to let trusted people in, and to let old hurts go, is where your power really opens up.",
		"With your Sun in Sagittarius, you are an explorer at heart. You love travel, learning, honest talk and big ideas, and your optimism lifts the people around you. Feeling free matters a lot to you. Your growth comes from following through on promises and finding adventure in commitment too.",
		"With your Sun in Capricorn, you are responsible, patient and ambitious. You set long-term goals and climb toward them steadily, and people rely on you. You often grow more relaxed and playful with age. Remembering that rest is part of success, not a reward for it, is your lesson.",
		"With your Sun in Aquarius, you think for yourself. You are friendly, original and interested in how things could be fairer or better for everyone. You need room to be different. Your growth comes from letting people close and sharing feelings as easily as you share ideas.",
		"With your Sun in Pisces, you are gentle, imaginative and deeply compassionate. You sense moods and needs before anyone says a word, and you are drawn to music, art and anything with soul. You need quiet time to recharge. Clear boundaries help your sensitivity become a strength instead of a drain.",
	},
	"moon": {
		"With your Moon in Aries, your feelings arrive fast and leave fast. You react in the moment, need to blow off steam physically and feel better once you have done something about a problem. You hate feeling trapped. A short pause before reacting saves you a lot of apologising.",
		"With your Moon in Taurus, you need calm, comfort and predictability to feel safe. Good food, nature, touch and familiar routines settle you quickly. Your moods are steady and slow to change. When you are upset, time and patience help you more than talking it through.",
		"With your Moon in Gemini, you process feelings by talking and thinking about them. You need mental stimulation, conversation and a bit of variety to feel at ease. Boredom unsettles you more than conflict. Writing things down helps you understand what you actually feel.",
		"With your Moon in Cancer, your feelings run deep and you are naturally nurturing. You need a safe home and close people around you, and you sense the mood in a room instantly. Old memories stay with you. Looking after your own needs first keeps you from running on empty.",
		"With your Moon in Leo, you need to feel seen, appreciated and loved out loud. You are warm-hearted and generous with your affection, and a little praise lifts you right up. Feeling ignored hurts. Creative play and fun are real medicine for your mood.",
		"With your Moon in Virgo, you feel calm when life is in order. Tidying, planning and being useful help you handle stress, and you show care through practical help. You can worry or criticise yourself when you are anxious. Remember that you do not need to earn rest.",
		"With your Moon in Libra, you feel best when your relationships are in harmony. You need beauty, fairness and company, and conflict upsets you more than you show. You are good at seeing how others feel. Make sure you also ask yourself what you want.",
		"With your Moon in Scorpio, you feel everything intensely, even if you show very little. You need deep trust and honesty, and you can sense when something is being hidden. You protect yourself carefully. Letting someone see your softer side builds the closeness you crave.",
		"With your Moon in Sagittarius, you need freedom, space and something to look forward to. You cheer yourself up with plans, travel, laughter and big ideas. You dislike heavy moods and too many rules. Sitting with a difficult feeling, instead of rushing past it, helps it pass for good.",
		"With your Moon in Capricorn, you handle feelings with self-control and keep going when things get tough. You feel safe when you have a plan and when you are making progress. You may find it hard to ask for help. Letting yourself lean on others is a strength, not a failure.",
		"With your Moon in Aquarius, you need space to be yourself and friends who accept you as you are. You often process feelings by stepping back and thinking about them. You feel best when you are part of something bigger. Remember that your feelings deserve as much attention as your ideas.",
		"With your Moon in Pisces, you are tender, intuitive and easily moved. You absorb the moods of people around you, so you need quiet time, music, water or sleep to reset. You are kind to a fault. Gentle boundaries keep your compassion from wearing you out.",
	},
	"asc": {
		"With Aries rising, you meet the world head-on. People see you as energetic, direct and ready for action, and you often take the lead without meaning to. You move fast and get bored by waiting. Your first impression is bold, so it helps to show your warmth too.",
		"With Taurus rising, you come across as calm, steady and easy to be around. People find you reassuring and often notice your good taste or relaxed presence. You prefer to take your time with new situations. Once you settle in, you are hard to move.",
		"With Gemini rising, you come across as bright, curious and talkative. You ask questions, make jokes and adapt quickly to whoever you are with. People see you as youthful at any age. You may seem restless, so show people the depth behind the quick wit.",
		"With Cancer rising, you come across as warm, gentle and a little protective. You size up whether a place feels safe before you open up. People often feel looked after by you. Your shell is real, but so is the softness behind it.",
		"With Leo rising, you make an entrance even when you do not try. People see you as warm, confident and generous, and they notice your style. You like to make a good impression. Letting people see your less polished side makes you even easier to love.",
		"With Virgo rising, you come across as modest, attentive and capable. You notice details, offer help and like to understand a situation before you jump in. People trust you to get things right. Try not to hide your warmth behind usefulness.",
		"With Libra rising, you come across as charming, polite and easy to like. You smooth over awkward moments and make people feel welcome. Looks and atmosphere matter to you. Make sure your pleasantness does not hide what you really think.",
		"With Scorpio rising, you come across as intense, private and quietly magnetic. People sense there is more to you than you show, and you notice a lot without saying much. You trust slowly. When you do open up, people feel the depth of your loyalty.",
		"With Sagittarius rising, you come across as friendly, open and full of plans. People see you as optimistic and fun, and you are often on the move. You say what you think. A little tact makes your honesty land as the gift it is meant to be.",
		"With Capricorn rising, you come across as composed, capable and serious. People see you as someone in charge, even when you are young. You prefer to be prepared and to prove yourself through results. Over time you relax and your dry humour shows.",
		"With Aquarius rising, you come across as friendly, independent and a little different. People notice your original style or ideas. You treat everyone as an equal and dislike being boxed in. Letting people past the friendly but cool surface brings real closeness.",
		"With Pisces rising, you come across as gentle, dreamy and kind. People feel they can confide in you, and you pick up on atmospheres immediately. You can blend in with whoever you are with. Knowing what you want keeps you from drifting into other people's plans.",
	},
	"mercury": {
		"With Mercury in Aries, you think fast and speak your mind. You make decisions quickly and get straight to the point, which can be refreshing or blunt. You learn best by doing. Counting to three before you reply helps your words land well.",
		"With Mercury in Taurus, you think carefully and say what you mean. You learn slowly but remember for good, and once you form an opinion you stick to it. You prefer practical ideas you can use. Staying open to new information keeps your thinking from getting stuck.",
		"With Mercury in Gemini, your mind is quick, witty and endlessly curious. You love to talk, read and juggle several ideas at once, and you explain things well. You can get scattered. Finishing one thought before starting the next makes your brilliance count.",
		"With Mercury in Cancer, you think with your feelings and remember things by how they felt. You are a sensitive listener and say things with care. Your memory for personal details is strong. Try to separate facts from feelings when a decision really matters.",
		"With Mercury in Leo, you speak with warmth and confidence and you love a good story. You make ideas exciting and people listen when you talk. You like your ideas to be appreciated. Inviting other people's views makes you an even better communicator.",
		"With Mercury in Virgo, your mind is precise, analytical and practical. You notice details, spot mistakes and like to understand how things work. You are a natural editor and problem-solver. Watch that careful thinking does not turn into worry.",
		"With Mercury in Libra, you think in terms of balance and fairness. You are tactful, persuasive and good at seeing both sides of an argument. You often think best in conversation. Deciding without waiting for everyone's approval is your growing edge.",
		"With Mercury in Scorpio, your mind digs deep. You notice hidden motives, ask probing questions and keep secrets well. You prefer meaningful conversations to small talk. Your words can be sharp, so use that insight to help rather than to wound.",
		"With Mercury in Sagittarius, you think in big pictures and love ideas about meaning, travel and possibility. You are honest, funny and enthusiastic when you talk. You can skip the details. Checking the small print keeps your big plans on track.",
		"With Mercury in Capricorn, you think practically and plan ahead. You speak with authority, choose your words carefully and prefer facts to speculation. Your dry humour often surprises people. Leaving room for imagination keeps your thinking fresh.",
		"With Mercury in Aquarius, your mind is original and independent. You enjoy new ideas, technology and conversations about how things could change. You can be stubborn once convinced. Listening to feelings as well as logic makes your ideas more persuasive.",
		"With Mercury in Pisces, you think in images, impressions and feelings more than in lists. You are imaginative, intuitive and good at sensing what is not being said. Details can slip away. Writing things down helps you turn inspiration into something real.",
	},
	"venus": {
		"With Venus in Aries, you love boldly and like to make the first move. Attraction for you is quick and exciting, and you enjoy a bit of challenge. You show love with enthusiasm and honesty. Keeping the spark alive once the chase is over is your lesson in love.",
		"With Venus in Taurus, you love with patience, loyalty and lots of physical affection. You enjoy comfort, beauty and simple pleasures shared with someone you trust. You take your time to commit. Watch that the wish for security does not become possessiveness.",
		"With Venus in Gemini, you fall for minds and conversation. You want a partner who is fun, curious and easy to talk to, and you flirt with words. You enjoy variety in what you do together. Showing your feelings, not just your wit, deepens your bonds.",
		"With Venus in Cancer, you love tenderly and protectively. You want closeness, emotional safety and a home to share, and you show love by caring for people. You can be shy at first. Saying what you need helps you get the care you give so freely.",
		"With Venus in Leo, you love generously and romantically. You enjoy grand gestures, fun dates and being openly adored, and you are just as generous in return. Loyalty matters a lot to you. Remember that quiet everyday affection counts too.",
		"With Venus in Virgo, you show love through practical care and attention to detail. You notice what people need and quietly provide it. You are selective and take time to trust. Letting go of the wish for a perfect partner makes room for a real one.",
		"With Venus in Libra, you are a natural romantic who values harmony, beauty and partnership. You are charming and considerate, and you enjoy doing things as a couple. You dislike conflict. Speaking up early about small issues keeps your relationships truly balanced.",
		"With Venus in Scorpio, you love deeply, intensely and all the way. You want a bond built on total trust and you are extremely loyal. Casual connections rarely satisfy you. Watch that jealousy does not get in the way of the closeness you want.",
		"With Venus in Sagittarius, you love freely and with good humour. You are attracted to people who broaden your world, and you enjoy travel and adventure together. You need room to breathe. Showing up consistently proves your love more than big promises.",
		"With Venus in Capricorn, you take love seriously and show it through commitment and effort. You value loyalty, reliability and shared goals, and you are often attracted to maturity. You can seem reserved. Letting your softer, playful side show strengthens your relationships.",
		"With Venus in Aquarius, you want a partner who is also your best friend. You value freedom, honesty and shared ideas, and you are drawn to people who are a little unusual. Too much clinginess makes you pull away. Showing affection openly helps others feel your loyalty.",
		"With Venus in Pisces, you love compassionately and romantically. You see the best in people and are willing to give a great deal. Music, art and dreamy dates move you. Keeping your eyes open to who someone really is protects your big heart.",
	},
	"mars": {
		"With Mars in Aries, your energy is fast, direct and competitive. You go after what you want without hesitation and you rarely back down from a challenge. Anger flares quickly and passes quickly. Exercise and clear goals are the best outlets for your fire.",
		"With Mars in Taurus, your energy is slow to start but almost impossible to stop. You work steadily and patiently toward what you want. You are hard to provoke, but formidable once you are. Remember to change course when a plan is no longer working.",
		"With Mars in Gemini, your energy goes into ideas, words and many projects at once. You argue well and enjoy a lively debate. Your attention jumps around. Focusing on one goal at a time helps you get much more done.",
		"With Mars in Cancer, you act from your feelings and fight hardest for the people you love. You may avoid direct confrontation and show anger indirectly. You are deeply protective. Saying clearly what bothers you, early on, clears the air.",
		"With Mars in Leo, your energy is warm, proud and dramatic. You want to lead, create and be recognised for what you do. You put your whole heart into your goals. Choosing your battles keeps pride from getting in the way.",
		"With Mars in Virgo, your energy is precise and hardworking. You get things done through careful planning and attention to detail, and you are good at fixing problems. Frustration can turn into criticism. Remember to rest; not everything needs improving today.",
		"With Mars in Libra, you prefer to act together with others and dislike open conflict. You fight for fairness, often on someone else's behalf. You can hesitate when making decisions. Asserting your own wishes is not unfair; it is part of balance.",
		"With Mars in Scorpio, your drive is intense, focused and private. You pursue your goals with determination and rarely give up. You keep anger inside until it is time to act. Channelling that intensity into a meaningful goal is your superpower.",
		"With Mars in Sagittarius, your energy is enthusiastic and adventurous. You are motivated by big goals, freedom and the chance to explore. You can start strong and lose interest. Pacing yourself helps you reach the finish line.",
		"With Mars in Capricorn, your drive is disciplined and ambitious. You set long-term goals and work toward them with patience and self-control. You are at your best with a clear plan. Make time for fun so ambition does not become all you do.",
		"With Mars in Aquarius, your energy goes into causes, ideas and doing things your own way. You work well in groups and fight for what is fair. You resist being told what to do. Turning rebellion into constructive change is where you shine.",
		"With Mars in Pisces, your energy moves in waves, guided by intuition and feeling. You are motivated by compassion and creativity more than competition. You may avoid conflict and get tired easily. Gentle routines and creative outlets keep your energy flowing.",
	},
	"jupiter": {
		"Jupiter in Aries brings luck through courage and initiative. You grow most when you take the first step and back yourself.",
		"Jupiter in Taurus brings growth through patience, practical skills and enjoying life's comforts. Steady effort tends to pay off well for you.",
		"Jupiter in Gemini brings opportunity through learning, writing and talking to many people. Your curiosity opens doors.",
		"Jupiter in Cancer brings blessings through home, family and caring for others. Generosity and kindness come back to you.",
		"Jupiter in Leo brings growth through creativity, generosity and confidence. You are lucky when you let yourself shine.",
		"Jupiter in Virgo brings opportunity through useful work, skill and attention to detail. Helping others helps you grow.",
		"Jupiter in Libra brings luck through partnership, fairness and good relationships. Cooperation multiplies your chances.",
		"Jupiter in Scorpio brings growth through depth, research and facing hard truths. You gain strength from what others avoid.",
		"Jupiter in Sagittarius brings luck through travel, study and an open mind. Your optimism is genuinely contagious.",
		"Jupiter in Capricorn brings success through discipline, patience and long-term planning. Your effort builds lasting rewards.",
		"Jupiter in Aquarius brings opportunity through friends, communities and new ideas. You grow by thinking differently.",
		"Jupiter in Pisces brings grace through compassion, imagination and faith. Kindness and creativity open the way for you.",
	},
	"saturn": {
		"Saturn in Aries asks you to learn patience and self-discipline in how you act. Over time you become a calm, steady leader.",
		"Saturn in Taurus asks you to build security slowly and carefully. Financial and practical wisdom grows with age.",
		"Saturn in Gemini asks you to take your mind and words seriously. You can become a careful, trusted communicator.",
		"Saturn in Cancer asks you to build emotional security from within. Family lessons shape you, and you become a strong support for others.",
		"Saturn in Leo asks you to earn your own confidence. Creative work done with discipline becomes something you are proud of.",
		"Saturn in Virgo asks you to master skills and healthy routines without becoming a perfectionist. You become highly capable over time.",
		"Saturn in Libra asks you to take relationships seriously and build fair, lasting commitments. Partnership is where you mature most.",
		"Saturn in Scorpio asks you to face fears around trust and control. Working through them gives you rare emotional strength.",
		"Saturn in Sagittarius asks you to build beliefs you have tested yourself. You grow into a wise, grounded teacher.",
		"Saturn in Capricorn gives a strong sense of duty and ambition. You build lasting achievements through steady work.",
		"Saturn in Aquarius asks you to find your place in groups while staying true to yourself. You become a responsible reformer.",
		"Saturn in Pisces asks you to give your dreams and compassion a practical shape. Clear boundaries make your kindness sustainable.",
	},
	"uranus": {
		"Uranus in Aries marks a generation of pioneers who want to start things fresh. You may feel a strong urge to act independently.",
		"Uranus in Taurus marks a generation that changes how we think about money, food and the planet. You may find new ways to feel secure.",
		"Uranus in Gemini marks a generation of new ways to learn and communicate. Your thinking can be strikingly original.",
		"Uranus in Cancer marks a generation that rethinks home and family. You may build a home life that looks unlike your parents'.",
		"Uranus in Leo marks a generation that expresses itself in bold, unusual ways. Your creativity has a rebellious streak.",
		"Uranus in Virgo marks a generation that changes work, health and technology. You may find unusual ways to be useful.",
		"Uranus in Libra marks a generation that redefines relationships. You may need a partnership that leaves room for freedom.",
		"Uranus in Scorpio marks a generation that brings hidden issues into the open. You may be drawn to deep, sudden transformations.",
		"Uranus in Sagittarius marks a generation that questions old beliefs. You may find meaning in unexpected places.",
		"Uranus in Capricorn marks a generation that reforms institutions and structures. You may want to change the system from within.",
		"Uranus in Aquarius marks a generation of inventors and networkers. Technology and community feel natural to you.",
		"Uranus in Pisces marks a generation with a fresh take on spirituality and art. Your intuition can come in sudden flashes.",
	},
	"neptune": {
		"Neptune in Aries marks a generation that dreams of courage and new beginnings. Your ideals push you to act.",
		"Neptune in Taurus marks a generation that idealises nature, beauty and simplicity. You may find peace in the physical world.",
		"Neptune in Gemini marks a generation that dreams through words and stories. Your imagination lives in language.",
		"Neptune in Cancer marks a generation that idealises home and family. You may long for a deep sense of belonging.",
		"Neptune in Leo marks a generation that dreams of glamour and creative expression. You may find magic in art and play.",
		"Neptune in Virgo marks a generation that idealises service and healing. You may feel called to help in practical ways.",
		"Neptune in Libra marks a generation that dreams of perfect love and peace. You may idealise partnership.",
		"Neptune in Scorpio marks a generation drawn to mystery, psychology and what lies beneath. You may sense hidden currents easily.",
		"Neptune in Sagittarius marks a generation that seeks meaning through travel, faith and philosophy. You may dream of a better world.",
		"Neptune in Capricorn marks a generation that questions authority and dreams of structures that actually serve people. Your ideals are practical.",
		"Neptune in Aquarius marks a generation that dreams of community and a shared future. You may idealise friendship and progress.",
		"Neptune in Pisces marks a generation of deep sensitivity and compassion. Art, music and spirituality speak to you strongly.",
	},
	"pluto": {
		"Pluto in Aries marks a generation driven to remake the world through bold action. You may find power in independence.",
		"Pluto in Taurus marks a generation that transforms how we value money, land and resources. You may find power in steadiness.",
		"Pluto in Gemini marks a generation that transforms information and communication. Words carry real power for you.",
		"Pluto in Cancer marks a generation shaped by deep changes in home, family and nation. Emotional roots run strong in you.",
		"Pluto in Leo marks a generation that transformed self-expression and individuality. You may find power in creativity.",
		"Pluto in Virgo marks a generation that transformed work, health and technology. You may find power in skill and analysis.",
		"Pluto in Libra marks a generation that transformed relationships and ideas of fairness. Partnership is where you meet your power.",
		"Pluto in Scorpio marks a generation that faces taboos and hidden truths head-on. You may be unusually resilient.",
		"Pluto in Sagittarius marks a generation that transforms beliefs, travel and global connection. You may find power in conviction.",
		"Pluto in Capricorn marks a generation that rebuilds institutions and authority. You may find power in long-term ambition.",
		"Pluto in Aquarius marks a generation that transforms technology and community. You may find power in collective change.",
		"Pluto in Pisces marks a generation that transforms spirituality and compassion. You may find power in imagination and empathy.",
	},
	"node": {
		"With your North Node in Aries, you grow by learning to stand on your own and act for yourself. Trusting your instincts brings fulfilment.",
		"With your North Node in Taurus, you grow by slowing down and building simple, steady security. Peace and patience bring fulfilment.",
		"With your North Node in Gemini, you grow by staying curious, asking questions and listening. Everyday conversations bring fulfilment.",
		"With your North Node in Cancer, you grow by opening up emotionally and investing in home and family. Letting yourself be cared for brings fulfilment.",
		"With your North Node in Leo, you grow by stepping into the spotlight and expressing yourself. Following your heart brings fulfilment.",
		"With your North Node in Virgo, you grow through practical service and healthy routines. Small, useful daily steps bring fulfilment.",
		"With your North Node in Libra, you grow by learning cooperation and true partnership. Meeting others halfway brings fulfilment.",
		"With your North Node in Scorpio, you grow by sharing deeply and letting things change. Trust and intimacy bring fulfilment.",
		"With your North Node in Sagittarius, you grow by exploring, learning and finding your own beliefs. The bigger picture brings fulfilment.",
		"With your North Node in Capricorn, you grow by taking responsibility and setting long-term goals. Building something of your own brings fulfilment.",
		"With your North Node in Aquarius, you grow by joining communities and working toward shared hopes. Being part of something larger brings fulfilment.",
		"With your North Node in Pisces, you grow by trusting, letting go and following your intuition. Compassion and quiet time bring fulfilment.",
	},
	"mc": {
		"With your Midheaven in Aries, you want a career where you can lead, compete and start new things. You are known for courage and initiative.",
		"With your Midheaven in Taurus, you want steady, meaningful work that builds lasting value. You are known for reliability and good taste.",
		"With your Midheaven in Gemini, you want varied work full of ideas, people and communication. You are known for your wit and versatility.",
		"With your Midheaven in Cancer, you want work that cares for people or feels like family. You are known for being supportive and protective.",
		"With your Midheaven in Leo, you want work where you can create, lead or perform. You are known for warmth and presence.",
		"With your Midheaven in Virgo, you want work that is useful, skilled and done well. You are known for competence and care.",
		"With your Midheaven in Libra, you want work involving people, beauty or fairness. You are known for diplomacy and charm.",
		"With your Midheaven in Scorpio, you want work with depth, focus and real stakes. You are known for determination and insight.",
		"With your Midheaven in Sagittarius, you want work with freedom, learning and purpose. You are known for enthusiasm and vision.",
		"With your Midheaven in Capricorn, you want to build a respected career step by step. You are known for responsibility and ambition.",
		"With your Midheaven in Aquarius, you want work that is innovative or serves the common good. You are known for original thinking.",
		"With your Midheaven in Pisces, you want work that is creative, healing or deeply meaningful. You are known for compassion and imagination.",
	},
}

// PlanetInSign describes a body or point in a sign. Unknown keys return "".
func PlanetInSign(planet, sign string) string {
	i, ok := signIndex[sign]
	if !ok {
		return ""
	}
	if t, ok := inSign[planet]; ok {
		return t[i]
	}
	q := signQuality[i]
	n := signs[i].Name
	switch planet {
	case "dsc":
		return "With your Descendant in " + n + ", you are drawn to partners who bring " + q + ". Through them you meet qualities you are still growing into yourself."
	case "ic":
		return "With your IC in " + n + ", home feels right when it has room for " + q + ". Your roots shape you more than you may notice."
	case "fortune":
		return "With your Part of Fortune in " + n + ", contentment comes through " + q + ". Lean into these when life feels flat."
	case "southnode":
		return "With your South Node in " + n + ", " + q + " already come naturally to you. Use these skills, but do not hide in them; your North Node shows where to grow."
	}
	return ""
}

// signQuality is a hand-written noun phrase per sign, used by the points
// that are read through what they bring rather than a full paragraph.
var signQuality = [12]string{
	"courage, independence and a spirit of adventure",
	"calm, comfort and steady loyalty",
	"curiosity, conversation and variety",
	"tenderness, care and a sense of family",
	"warmth, play and generous self-expression",
	"order, usefulness and careful attention",
	"harmony, beauty and fair give-and-take",
	"depth, trust and emotional honesty",
	"freedom, humour and a sense of meaning",
	"structure, ambition and reliability",
	"friendship, originality and the freedom to be yourself",
	"compassion, imagination and quiet time",
}
