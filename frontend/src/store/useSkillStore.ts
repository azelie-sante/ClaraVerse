import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import * as skillService from '@/services/skillService';
import type {
  Skill,
  UserSkillWithDetails,
  CreateSkillRequest,
  CommunitySkillEntry,
} from '@/services/skillService';

interface SkillState {
  // Data
  skills: Skill[];
  categories: Record<string, Skill[]>;
  userSkills: UserSkillWithDetails[];
  enabledSkillIds: Set<string>;

  // Community
  communitySkills: CommunitySkillEntry[];
  communityLoading: boolean;
  communityError: string | null;

  // UI
  loading: boolean;
  error: string | null;

  // Actions
  fetchSkills: (category?: string) => Promise<void>;
  fetchMySkills: () => Promise<void>;
  getSkill: (id: string) => Promise<Skill>;
  enableSkill: (skillId: string) => Promise<void>;
  disableSkill: (skillId: string) => Promise<void>;
  createSkill: (data: CreateSkillRequest) => Promise<Skill>;
  updateSkill: (id: string, data: CreateSkillRequest) => Promise<void>;
  deleteSkill: (id: string) => Promise<void>;
  isSkillEnabled: (skillId: string) => boolean;

  // Import / Export / Community
  fetchCommunitySkills: () => Promise<void>;
  importFromSkillMD: (content: string) => Promise<Skill>;
  importFromGitHub: (url: string) => Promise<Skill>;
  exportSkillMD: (id: string) => Promise<string>;
}

export const useSkillStore = create<SkillState>()(
  devtools(
    (set, get) => ({
      skills: [],
      categories: {},
      userSkills: [],
      enabledSkillIds: new Set<string>(),
      communitySkills: [],
      communityLoading: false,
      communityError: null,
      loading: false,
      error: null,

      fetchSkills: async (category?: string) => {
        set({ loading: true, error: null });
        try {
          const response = await skillService.listSkills(category);
          set({
            skills: response.skills ?? [],
            categories: response.categories ?? {},
            loading: false,
          });
        } catch (err) {
          const message = err instanceof Error ? err.message : 'Failed to fetch skills';
          set({ loading: false, error: message });
        }
      },

      fetchMySkills: async () => {
        try {
          const response = await skillService.getMySkills();
          const userSkills = response.skills ?? [];
          const enabledIds = new Set<string>();
          for (const us of userSkills) {
            if (us.enabled) {
              enabledIds.add(us.skill_id);
            }
          }
          set({ userSkills, enabledSkillIds: enabledIds });
        } catch (err) {
          const message = err instanceof Error ? err.message : 'Failed to fetch user skills';
          set({ error: message });
        }
      },

      getSkill: async (id: string) => {
        return skillService.getSkill(id);
      },

      enableSkill: async (skillId: string) => {
        await skillService.enableSkill(skillId);
        const newSet = new Set(get().enabledSkillIds);
        newSet.add(skillId);
        set({ enabledSkillIds: newSet });
      },

      disableSkill: async (skillId: string) => {
        await skillService.disableSkill(skillId);
        const newSet = new Set(get().enabledSkillIds);
        newSet.delete(skillId);
        set({ enabledSkillIds: newSet });
      },

      createSkill: async (data: CreateSkillRequest) => {
        const skill = await skillService.createSkill(data);
        set({ skills: [...get().skills, skill] });
        return skill;
      },

      updateSkill: async (id: string, data: CreateSkillRequest) => {
        await skillService.updateSkill(id, data);
        set({
          skills: get().skills.map(s =>
            s.id === id ? { ...s, ...data, updated_at: new Date().toISOString() } : s
          ),
        });
      },

      deleteSkill: async (id: string) => {
        await skillService.deleteSkill(id);
        set({ skills: get().skills.filter(s => s.id !== id) });
      },

      isSkillEnabled: (skillId: string) => {
        return get().enabledSkillIds.has(skillId);
      },

      fetchCommunitySkills: async () => {
        set({ communityLoading: true, communityError: null });
        try {
          const response = await skillService.listCommunitySkills();
          set({
            communitySkills: response.skills ?? [],
            communityLoading: false,
          });
        } catch (err) {
          const message = err instanceof Error ? err.message : 'Failed to fetch community skills';
          set({ communityLoading: false, communityError: message });
        }
      },

      importFromSkillMD: async (content: string) => {
        const skill = await skillService.importFromSkillMD(content);
        set({ skills: [...get().skills, skill] });
        return skill;
      },

      importFromGitHub: async (url: string) => {
        const skill = await skillService.importFromGitHub(url);
        set({ skills: [...get().skills, skill] });
        return skill;
      },

      exportSkillMD: async (id: string) => {
        return skillService.exportSkillMD(id);
      },
    }),
    { name: 'skill-store' }
  )
);
