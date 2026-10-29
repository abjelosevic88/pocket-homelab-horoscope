package content

// inHouse holds hand-written texts for each body in each house (1–12).
var inHouse = map[string][12]string{
	"sun": {
		"With your Sun in the 1st house, your personality is front and centre. People see who you are straight away, and you grow by being openly yourself.",
		"With your Sun in the 2nd house, you find yourself through what you build and earn. Self-worth and financial security are central themes in your life.",
		"With your Sun in the 3rd house, you shine through talking, writing and learning. Siblings, neighbours and everyday connections matter a lot to you.",
		"With your Sun in the 4th house, home and family are at the centre of your life. You recharge in private and often grow stronger as you get older.",
		"With your Sun in the 5th house, you come alive through creativity, play and romance. Children, hobbies and fun are where you shine.",
		"With your Sun in the 6th house, you find purpose in doing useful work well. Healthy routines and helping others make you feel like yourself.",
		"With your Sun in the 7th house, you discover yourself through close relationships. Partnership, in love or in work, is a major part of your path.",
		"With your Sun in the 8th house, you are drawn to depth, intimacy and life's big changes. You find yourself by going through transformation rather than avoiding it.",
		"With your Sun in the 9th house, you shine through travel, study and the search for meaning. Broadening your horizons keeps you alive.",
		"With your Sun in the 10th house, career and reputation matter a lot to you. You are meant to be visible and to achieve something you are proud of.",
		"With your Sun in the 11th house, you shine among friends and in groups. Shared hopes and causes give your life direction.",
		"With your Sun in the 12th house, you have a rich inner life and need time alone. You often shine most quietly, behind the scenes or in service of others.",
	},
	"moon": {
		"With your Moon in the 1st house, your feelings show on your face. You are sensitive to your surroundings and others pick up on your moods quickly.",
		"With your Moon in the 2nd house, you feel safe when your finances and possessions are in order. Comfort and familiar things soothe you.",
		"With your Moon in the 3rd house, you process feelings by talking and writing. A chat with a sibling or friend can change your whole mood.",
		"With your Moon in the 4th house, home is your emotional anchor. You need a private, cosy space and close family ties to feel settled.",
		"With your Moon in the 5th house, you feel best when you are having fun, creating or in love. Play is emotionally important for you.",
		"With your Moon in the 6th house, your mood is tied to your routines and your health. A tidy day and a useful task calm you down.",
		"With your Moon in the 7th house, you need a close partner to feel emotionally complete. You are very tuned in to the people you are with.",
		"With your Moon in the 8th house, your feelings run deep and private. You need real trust and emotional honesty to feel safe with someone.",
		"With your Moon in the 9th house, you feel free and happy when you are learning or travelling. A sense of meaning keeps your spirits up.",
		"With your Moon in the 10th house, your feelings are linked to your work and public image. You may work in a caring role or be visible to many people.",
		"With your Moon in the 11th house, friends are your emotional family. You feel good when you belong to a group with shared values.",
		"With your Moon in the 12th house, you keep many feelings to yourself. Solitude, sleep and quiet reflection are how you recharge.",
	},
	"mercury": {
		"With Mercury in the 1st house, you come across as quick, curious and talkative. People notice your mind before anything else.",
		"With Mercury in the 2nd house, you think practically about money and value. You may earn through words, ideas or advice.",
		"With Mercury in the 3rd house, communication is natural to you. You learn quickly and enjoy writing, reading and short trips.",
		"With Mercury in the 4th house, you do your best thinking at home. Family conversations and memories shape your mind.",
		"With Mercury in the 5th house, you think creatively and playfully. Games, stories and flirty banter come easily to you.",
		"With Mercury in the 6th house, you are an organised, detail-minded worker. You think best with a list and a clear task.",
		"With Mercury in the 7th house, you think best in conversation with others. Partners and colleagues sharpen your ideas.",
		"With Mercury in the 8th house, your mind digs deep. You are drawn to research, psychology and what lies beneath the surface.",
		"With Mercury in the 9th house, you love big ideas, languages and learning. Travel and study broaden your thinking.",
		"With Mercury in the 10th house, communication plays a big role in your career. You may be known for what you say or write.",
		"With Mercury in the 11th house, you enjoy exchanging ideas with friends and groups. Networks and online communities suit your mind.",
		"With Mercury in the 12th house, you think intuitively and keep many thoughts private. Writing in a journal helps you find clarity.",
	},
	"venus": {
		"With Venus in the 1st house, you come across as charming and likeable. You care about looks and enjoy making a pleasant impression.",
		"With Venus in the 2nd house, you enjoy beautiful things and have a good sense of value. Money can come to you through pleasant work.",
		"With Venus in the 3rd house, you speak kindly and diplomatically. You often have warm relationships with siblings and neighbours.",
		"With Venus in the 4th house, you love a beautiful, peaceful home. Family harmony matters a lot to you.",
		"With Venus in the 5th house, romance, art and fun come naturally. You enjoy dating, creating and celebrating.",
		"With Venus in the 6th house, you like a pleasant workplace and friendly colleagues. You show love through everyday help.",
		"With Venus in the 7th house, partnership is a source of happiness for you. You are a natural partner and often attract kind people.",
		"With Venus in the 8th house, you love deeply and intensely. Shared resources and emotional intimacy are important in your relationships.",
		"With Venus in the 9th house, you love travel, foreign cultures and ideas. You may fall for someone from a different background.",
		"With Venus in the 10th house, charm helps your career. You may work in beauty, art, diplomacy or with the public.",
		"With Venus in the 11th house, friendships bring you joy. You are popular in groups and may meet partners through friends.",
		"With Venus in the 12th house, you love quietly and selflessly. Private romance and creative solitude mean a lot to you.",
	},
	"mars": {
		"With Mars in the 1st house, you come across as energetic and bold. You act quickly and like to take the lead.",
		"With Mars in the 2nd house, you work hard for your money and want to earn it yourself. You defend what is yours.",
		"With Mars in the 3rd house, you speak with energy and enjoy debate. You are always busy with errands and ideas.",
		"With Mars in the 4th house, you put a lot of energy into your home. Family life can be lively, and you may love DIY projects.",
		"With Mars in the 5th house, you pour energy into play, sport, romance and creative projects. You are competitive in fun.",
		"With Mars in the 6th house, you are a tireless worker who likes to stay active. Exercise is especially good for your mood.",
		"With Mars in the 7th house, your relationships are lively and passionate. You may attract strong-willed partners.",
		"With Mars in the 8th house, you have intense desires and great inner strength. You handle crises with courage.",
		"With Mars in the 9th house, you are passionate about your beliefs and love adventure. Travel and study energise you.",
		"With Mars in the 10th house, you are ambitious and driven in your career. You want to achieve and be seen achieving.",
		"With Mars in the 11th house, you fight for causes and push groups into action. Friends find you motivating.",
		"With Mars in the 12th house, you work best behind the scenes. Quiet exercise and private time help you release tension.",
	},
	"jupiter": {
		"With Jupiter in the 1st house, you come across as warm, optimistic and generous. Opportunities often come when you show yourself.",
		"With Jupiter in the 2nd house, you have a knack for attracting resources. Generosity comes naturally, but watch the spending.",
		"With Jupiter in the 3rd house, you learn easily and communicate with enthusiasm. Your curiosity brings opportunities.",
		"With Jupiter in the 4th house, home is a place of growth and plenty. A big, welcoming home or family often matters to you.",
		"With Jupiter in the 5th house, you find luck through creativity, play and romance. Joy comes easily to you.",
		"With Jupiter in the 6th house, you do well at work and enjoy being helpful. Healthy habits pay off generously for you.",
		"With Jupiter in the 7th house, partners bring growth and good fortune. You tend to meet generous, helpful people.",
		"With Jupiter in the 8th house, you gain through shared resources and deep bonds. You come out of hard times stronger.",
		"With Jupiter in the 9th house, travel, study and philosophy bring you luck. Your optimism inspires others.",
		"With Jupiter in the 10th house, career opportunities come naturally. You can earn a good reputation through generosity and vision.",
		"With Jupiter in the 11th house, friends and groups bring you luck. Your network can open many doors.",
		"With Jupiter in the 12th house, you have a quiet inner protection. Rest, reflection and kindness bring unexpected help.",
	},
	"saturn": {
		"With Saturn in the 1st house, you come across as serious and responsible. You grow more confident and relaxed with time.",
		"With Saturn in the 2nd house, you learn careful habits with money. Security comes slowly, but it lasts.",
		"With Saturn in the 3rd house, you think carefully before you speak. You may have worked hard to find your voice.",
		"With Saturn in the 4th house, home and family carry responsibility for you. You build your own sense of security over time.",
		"With Saturn in the 5th house, you take creativity and fun seriously. Letting yourself play is part of your growth.",
		"With Saturn in the 6th house, you are a dedicated, reliable worker. Good health habits are important for you.",
		"With Saturn in the 7th house, you take relationships seriously and may commit later in life. Lasting partnerships matter to you.",
		"With Saturn in the 8th house, trust takes time for you. Learning to share and let go is a lifelong lesson.",
		"With Saturn in the 9th house, you take beliefs and education seriously. You build your own philosophy through experience.",
		"With Saturn in the 10th house, you are ambitious and work hard for recognition. Success often comes later, but it lasts.",
		"With Saturn in the 11th house, you prefer a few loyal friends to a big crowd. You may take on responsibility in groups.",
		"With Saturn in the 12th house, you face private fears and need time alone. Quiet discipline helps you find inner peace.",
	},
	"uranus": {
		"With Uranus in the 1st house, you come across as unusual and independent. You need freedom to be yourself.",
		"With Uranus in the 2nd house, your income may go up and down. Unusual ways of earning can suit you.",
		"With Uranus in the 3rd house, your mind is quick and original. You may learn in unconventional ways.",
		"With Uranus in the 4th house, home life may have been unpredictable. You need a home that gives you freedom.",
		"With Uranus in the 5th house, your creativity is original and surprising. Romance may arrive suddenly.",
		"With Uranus in the 6th house, you need variety and freedom at work. Flexible routines suit you better than rigid ones.",
		"With Uranus in the 7th house, you need a partner who gives you space. Relationships may start or change suddenly.",
		"With Uranus in the 8th house, deep changes can come unexpectedly. You may have sudden insights into yourself.",
		"With Uranus in the 9th house, you think independently about beliefs and enjoy unusual travels. You question everything.",
		"With Uranus in the 10th house, your career may take surprising turns. You do best where you can innovate.",
		"With Uranus in the 11th house, you have unusual friends and love groups with new ideas. Community and technology suit you.",
		"With Uranus in the 12th house, your intuition can arrive in sudden flashes. You may have an unconventional inner life.",
	},
	"neptune": {
		"With Neptune in the 1st house, you come across as gentle, dreamy or hard to pin down. You absorb the moods around you.",
		"With Neptune in the 2nd house, you may be idealistic or a little vague about money. Creative or caring work can bring income.",
		"With Neptune in the 3rd house, you think in images and have a vivid imagination. Poetry, music and stories come naturally.",
		"With Neptune in the 4th house, you may idealise home and family. You need a peaceful, gentle place to retreat.",
		"With Neptune in the 5th house, you are highly creative and romantic. Make sure love is based on who someone really is.",
		"With Neptune in the 6th house, you want work that helps or heals. Gentle routines support your sensitive system.",
		"With Neptune in the 7th house, you seek a soulmate and may idealise partners. Clear eyes protect your kind heart.",
		"With Neptune in the 8th house, you have strong intuition about hidden things. Clear boundaries matter with shared money.",
		"With Neptune in the 9th house, you are drawn to spirituality, art and faraway places. You dream of a better world.",
		"With Neptune in the 10th house, you want a career with meaning or creativity. Your path may take time to become clear.",
		"With Neptune in the 11th house, you have idealistic friendships and dreams for the future. You may support good causes.",
		"With Neptune in the 12th house, you have a rich inner world and deep compassion. Meditation, music and rest restore you.",
	},
	"pluto": {
		"With Pluto in the 1st house, you have a strong, intense presence. Your life involves several deep personal reinventions.",
		"With Pluto in the 2nd house, money and possessions can bring big lessons. You can rebuild your resources from scratch.",
		"With Pluto in the 3rd house, your words carry weight. You are drawn to investigating and getting to the bottom of things.",
		"With Pluto in the 4th house, family history runs deep in you. Transforming your roots is part of your path.",
		"With Pluto in the 5th house, you love and create intensely. Creative work can transform you.",
		"With Pluto in the 6th house, you can work with remarkable focus. Changing your habits changes your whole life.",
		"With Pluto in the 7th house, relationships are intense and transformative. You may attract powerful partners.",
		"With Pluto in the 8th house, you are drawn to life's depths and handle crises well. You are deeply resilient.",
		"With Pluto in the 9th house, your beliefs go through deep change. You search hard for truth.",
		"With Pluto in the 10th house, you are ambitious and may hold real influence. Your career can go through big transformations.",
		"With Pluto in the 11th house, friends and groups can change your life. You may lead or reshape communities.",
		"With Pluto in the 12th house, you have deep inner resources and intuition. Quiet inner work brings great strength.",
	},
	"node": {
		"With your North Node in the 1st house, you grow by putting yourself first sometimes and trusting your own choices.",
		"With your North Node in the 2nd house, you grow by building your own resources and learning what you truly value.",
		"With your North Node in the 3rd house, you grow by staying curious, asking questions and talking to the people around you.",
		"With your North Node in the 4th house, you grow by investing in home, family and your inner life.",
		"With your North Node in the 5th house, you grow by having fun, creating and letting your heart lead.",
		"With your North Node in the 6th house, you grow through practical work, healthy routines and service.",
		"With your North Node in the 7th house, you grow through partnership and learning to share decisions.",
		"With your North Node in the 8th house, you grow by sharing deeply, trusting others and letting things change.",
		"With your North Node in the 9th house, you grow by travelling, studying and finding your own beliefs.",
		"With your North Node in the 10th house, you grow by stepping into public life and building a career.",
		"With your North Node in the 11th house, you grow through friends, groups and working toward shared goals.",
		"With your North Node in the 12th house, you grow by slowing down, reflecting and trusting something larger than yourself.",
	},
}

// nodeHouseTail completes the North Node house texts (written as one sentence
// above) with a second sentence that points back at the comfort zone.
const nodeHouseTail = " It may feel unfamiliar at first, but this is where life feels most meaningful."

// PlanetInHouse describes a body in a house (1–12). Unknown keys return "".
func PlanetInHouse(planet string, house int) string {
	if house < 1 || house > 12 {
		return ""
	}
	t, ok := inHouse[planet]
	if !ok {
		return ""
	}
	if planet == "node" {
		return t[house-1] + nodeHouseTail
	}
	return t[house-1]
}
