those are the endpoints

tested - done - postLogout(): Promise<void>;
tested - done - postLogin(login: string, password: string): Promise<LoginResponse>;
tested - done - postSearchUsers(query: string): Promise<User[]>;
tested - done - postEditJournalItem(journalId: string, updatedItem: JournalItem): Promise<void>;
tested - done - postAddJournalItem(journalId: string, updatedItem: JournalItem): Promise<void>;
tested - done - postAddJournal(id: string, title: string): Promise<void>;
tested - done - postUpdateJournal(id: string, title: string): Promise<void>;
tested - done - done - postAddReadingUsersForJournal(journalId: string, userIds: string[]): Promise<void>;
tested - done - postReadingUsers(journalId: string): Promise<User[]>;
tested - done - postRemoveReading(journalId: string, userIds: string[]): Promise<void>;
tested - done - postJournalItems(journalId: string): Promise<JournalItem[]>;
tested - done - postRemoveJournals(selectedJournals: string[]): Promise<void>;
tested - done - postRemoveItems(selectedEntries: string[]): Promise<void>;

- []interface{} replace with []any

